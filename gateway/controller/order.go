package controller

import (
	"e-commerce/logger"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func CreateOrder(c *gin.Context) {
	var req models.Order
	userID, exists := c.Get("userID")
	if !exists {
		logger.Logger.Error("Unable to retrieve the userID from JWT token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		logger.Logger.Error("Invalid user ID format")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}
	req.UserID = userIDStr

	logger.Logger.Infof("Creating order for UserID=%s", userIDStr)
	orderID := uuid.New().ID()
	req.OrderID = orderID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	logger.Logger.Infof("Order request payload: %+v", req)

	// Connect to Inventory Service
	connInventory, err := grpc.Dial("inventory-service:50051", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to Inventory Service: %v", err)
	}
	defer connInventory.Close()
	inventoryClient := protobuf.NewInventoryServiceClient(connInventory)

	// Check product availability
	connDB, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to Write-DB Service: %v", err)
	}
	defer connDB.Close()
	productClient := protobuf.NewProductServiceClient(connDB)

	for _, prod := range req.OrderItems {
		grpcRequest := &protobuf.ProductIDRequest{
			ProductId: uint32(prod.ProductID),
		}

		_, err := productClient.GetProduct(c, grpcRequest)
		if err != nil {
			logger.Logger.Errorf("Failed to fetch product details for ProductID=%d: %v", prod.ProductID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch product details for ProductID=%d", prod.ProductID)})
			return
		}
	}

	// Check product stock availability
	var reserved int32
	for _, item := range req.OrderItems {
		grpcRequest := &protobuf.StockUpdateRequest{
			ProductId: uint32(item.ProductID),
			Quantity:  uint32(item.Quantity),
		}

		res, err := inventoryClient.UpdateStock(c, grpcRequest)
		if err != nil {
			logger.Logger.Errorf("Failed to update stock for ProductID=%d: %v", item.ProductID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}

		if res.Status == "Not Found" {
			logger.Logger.Warnf("Product not found for ProductID=%d", item.ProductID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		} else if res.Status == "Insufficient Stock" {
			logger.Logger.Warnf("Insufficient stock for ProductID=%d", item.ProductID)
			c.JSON(http.StatusConflict, gin.H{"error": "Insufficient stock"})
			return
		} else {
			logger.Logger.Infof("Stock updated successfully for ProductID=%d", item.ProductID)
			reserved = 1
		}
	}

	if reserved == 1 {
		// Create order in DB

		var items []*protobuf.OrderItem
		for _, item := range req.OrderItems {
			items = append(items, &protobuf.OrderItem{
				ProductId: item.ProductID,
				Quantity:  uint32(item.Quantity),
				Price:     float32(item.Price),
			})
		}

		orderRequest := &protobuf.OrderRequest{
			OrderId:     req.OrderID,
			UserId:      req.UserID,
			TotalAmount: float32(req.TotalPrice),
			Items:       items,
		}

		orderClient := protobuf.NewOrderServiceClient(connDB)
		response, err := orderClient.CreateOrder(c, orderRequest)
		if err != nil {
			logger.Logger.Errorf("Failed to create order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		logger.Logger.Infof("Order created successfully: %+v", response)
		c.JSON(http.StatusCreated, gin.H{"status": req.Status, "data": response})
	}
}

func GetOrder(c *gin.Context) {
	var orderID struct {
		OrderId uint32 `json:"order_id"`
	}
	if err := c.ShouldBindJSON(&orderID); err != nil {
		logger.Logger.Errorf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to Write-DB Service: %v", err)
	}
	defer conn.Close()

	orderClient := protobuf.NewOrderServiceClient(conn)
	grpcRequest := &protobuf.OrderIDRequest{OrderId: orderID.OrderId}

	response, err := orderClient.GetOrder(c, grpcRequest)
	if err != nil {
		logger.Logger.Errorf("Failed to fetch order details for OrderID=%d: %v", orderID.OrderId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order details"})
		return
	}

	logger.Logger.Infof("Order fetched successfully: %+v", response)
	c.JSON(http.StatusOK, gin.H{
		"order_id":     response.OrderId,
		"user_id":      response.UserId,
		"total_amount": response.TotalAmount,
		"items":        response.Items,
	})
}

func TestHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		logger.Logger.Error("User ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	logger.Logger.Infof("TestHandler called with UserID=%s", userID)
	ctx.JSON(http.StatusOK, gin.H{"userID": userID})
}
