package handler

import (
	"context"
	"e-commerce/logger"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var ProductCreationTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "product_creation_total",
		Help: "Total number of products created",
	},
)
var HttpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "endpoint", "status"},
)

func Init() {
	prometheus.MustRegister(ProductCreationTotal)
	prometheus.MustRegister(HttpRequestsTotal)
}

func CreateProduct(c *gin.Context) {
	ProductCreationTotal.Inc()
	logger.Logger.Info("Received CreateProduct request")

	var prod *models.Product

	if err := c.BindJSON(&prod); err != nil {
		logger.Logger.Errorf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request payload"})
		return
	}
	logger.Logger.Infof("Product payload: %+v", prod)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to write-db-service: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductRequest{
		ProductId: uint32(prod.ProductID), Name: prod.Name, Price: float32(prod.Price), Description: prod.Description, Quantity: uint32(prod.StockQuantity),
	}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.CreateProduct(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while creating product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("Product created successfully: %+v", response)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": response})
}

func GetProduct(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		logger.Logger.Errorf("Failed to convert ID to integer: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	logger.Logger.Infof("Received GetProduct request for ProductID=%d", id)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to write-db-service: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductIDRequest{ProductId: uint32(id)}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.GetProduct(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while fetching product: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("Product fetched successfully: %+v", response)
	ctx.JSON(http.StatusOK, response)
}

func GetAllProducts(ctx *gin.Context) {
	logger.Logger.Info("Received GetAllProducts request")

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to write-db-service: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.Empty{}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.ListProducts(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while fetching all products: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("All products fetched successfully")
	ctx.JSON(http.StatusOK, response)
}

func DeleteProduct(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		logger.Logger.Errorf("Failed to convert ID to integer: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	logger.Logger.Infof("Received DeleteProduct request for ProductID=%d", id)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to write-db-service: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductIDRequest{ProductId: uint32(id)}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.DeleteProduct(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while deleting product: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("Product deleted successfully: %+v", response)
	ctx.JSON(http.StatusOK, response)
}

func UpdateProduct(ctx *gin.Context) {
	logger.Logger.Info("Received UpdateProduct request")

	var prod *models.Product

	if err := ctx.BindJSON(&prod); err != nil {
		logger.Logger.Errorf("Failed to bind JSON: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request payload"})
		return
	}
	logger.Logger.Infof("Product payload: %+v", prod)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to write-db-service: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductRequest{ProductId: uint32(prod.ProductID), Name: prod.Name, Price: float32(prod.Price), Description: prod.Description}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.UpdateProduct(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while updating product: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("Product updated successfully: %+v", response)
	ctx.JSON(http.StatusOK, response)
}
