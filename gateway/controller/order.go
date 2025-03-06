package controller

import (
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func CreateOrder(c *gin.Context) {
	var req models.Order
	userID, errb := c.Get("userID")
	log.Println(userID, errb)

	if !errb {
		log.Println("Unable to retrieve the userid from jwt token")
		return
	}
	// Convert userID to string (if needed)
	user_id, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}
	// user_id, errs := strconv.Atoi(userIDStr)
	// if errs != nil {
	// 	log.Println("Unable to convert to integer")
	// }
	req.UserID = (user_id)

	// Create an order

	fmt.Println("Creating Order...")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Println(req.OrderItems)

	// Check stock availability
	// grpcRequestI := &protobuf.StockRequest{ProductId: (req.ID)}

	conn1, err2 := grpc.Dial(":50051", grpc.WithInsecure())
	if err2 != nil {
		log.Fatalf("did not connect to Inventory Service: %v", err2)
	}
	defer conn1.Close()
	client1 := protobuf.NewInventoryServiceClient(conn1)

	// res, err4 := client1.TrackStock(c, grpcRequestI)

	//Checking product availability

	conn, err := grpc.Dial(":50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	for _, prod := range req.OrderItems {

		grpcRequest := &protobuf.ProductIDRequest{
			ProductId: uint32(prod.ProductID),
		}
		client := protobuf.NewProductServiceClient(conn)

		_, err := client.GetProduct(c, grpcRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("\ngRPC error for fetching product id: %d: %v", prod.ProductID, err)})
			log.Printf("\nUnable to get product details from order items product %d", prod.ProductID)
			return
		}

	}

	//Checking product stock availability
	var reserved int32
	for _, val := range req.OrderItems {
		grpcRequestI := &protobuf.StockUpdateRequest{ProductId: (val.ProductID), Quantity: uint32(val.Quantity)}

		res, err := client1.UpdateStock(c, grpcRequestI)
		if err != nil {
			log.Println("Unable to update the stock", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err, "Reason": "Unable to update the stock"})
			return
		}

		// // Check if sufficient stock is available
		// if res.Status == "Insufficient Stock" {
		// 	// If out of stock, publish OutOfStock event
		// 	event := map[string]interface{}{
		// 		"product_id": val.ProductID,

		// 	}
		// 	eventJSON, err := json.Marshal(event)
		// 	if err != nil {
		// 		log.Println("Error marshalling OutOfStock event:", err)
		// 	} else {
		// 		s.kafkaClient.PublishMessage(c, &protobuf.PublishRequest{Topic: "Orders", EventType: "OutOfStock", Message: string(eventJSON)})
		// 	}

		// }
		// Send InventoryReserved event

		// var order models.Order

		// eventJSON, err := json.Marshal(&order)
		// if err != nil {
		// 	log.Println("Error marshalling InventoryReserved event:", err)
		// } else {
		// 	s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "InventoryReserved", Message: string(eventJSON)})
		// }
		// return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: res.Q, Status: "InventoryReserved"}, nil

		//Checking status
		if res.Status == "Not Found" {
			log.Printf("\nNo product foud with this product id: %v", val.ProductID)
			c.JSON(http.StatusInternalServerError, gin.H{"response": res})

		} else if res.Status == "Insufficient Stock" {
			log.Printf("\nInsufficient stock for this product id: %v", val.ProductID)
			c.JSON(http.StatusOK, res)
		} else {
			log.Printf("\n Updated stock for this product id: %v", val.ProductID)
			c.JSON(http.StatusOK, gin.H{"response": res, "msg": "Updated stock for this product"})
			reserved = 1
		}
		// conn1, err2 := grpc.Dial(":50052", grpc.WithInsecure())
		// if err2 != nil {
		// 	log.Fatalf("did not connect to producer: %v", err2)
		// }
		// defer conn1.Close()
		// grpcRequestK := &protobuf.PublishRequest{Topic: "Orders", EventType: "InventoryReserved", Message: ""}
		// client1 := protobuf.NewKafkaProducerServiceClient(conn1)
		// _, err := client1.PublishMessage(c, grpcRequestK)
		// if err4 != nil || res.Quantity < uint32(val.Quantity) {
		// 	c.JSON(http.StatusConflict, gin.H{"error": "Insufficient stock", "product": val})
		// 	return
		// }

		// stockResp, err := h.InventorySvc.TrackStock(context.Background(), &protobuf.StockRequest{ProductId: (val.Quantity)})
		// if err != nil || stockResp.Quantity < int32(val.Quantity) {
		// 	c.JSON(http.StatusConflict, gin.H{"error": "Insufficient stock"})
		// 	return
		// }

	}

	if reserved == 1 {

		// Create order in DB
		var order models.Order
		id := uuid.New().ID()
		order.OrderID = id
		order.UserID = req.UserID
		var items []*protobuf.OrderItem
		for _, item := range req.OrderItems {
			items = append(items, &protobuf.OrderItem{
				ProductId: item.ProductID,
				Quantity:  uint32(item.Quantity),
				Price:     float32(item.Price),
			})
		}
		log.Println(items)

		grpcRequest := &protobuf.OrderRequest{OrderId: order.OrderID, UserId: order.UserID, TotalAmount: float32(order.TotalPrice), Items: items}

		//Store in DB
		conn, err := grpc.Dial(":50001", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		log.Println("Connetected to DB")
		defer conn.Close()
		client := protobuf.NewOrderServiceClient(conn)

		response, err := client.CreateOrder(c, grpcRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
			log.Println("grpc error", err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "success", "data": response})

	}
}
