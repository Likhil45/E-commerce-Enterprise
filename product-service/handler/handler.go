package handler

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func CreateProduct(c *gin.Context) {

	var prod *models.Product

	if err := c.BindJSON(&prod); err != nil {
		log.Println("Unable to Bind JSon")
		log.Println(prod)
		c.JSON(http.StatusInternalServerError, err)

	}
	log.Println(prod)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcRequest := &protobuf.ProductRequest{
		ProductId: uint32(prod.ProductID), Name: prod.Name, Price: float32(prod.Price), Description: prod.Description, Quantity: uint32(prod.StockQuantity),
	}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.CreateProduct(context.Background(), grpcRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		log.Println("grpc error")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": response})
}

func GetProduct(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err1 := strconv.Atoi(idstr)
	if err1 != nil {
		log.Println("Unable to convert to integer")
	}
	log.Println(id)
	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductIDRequest{ProductId: uint32(id)}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.GetProduct(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, response)

}

func GetAllProducts(ctx *gin.Context) {

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.Empty{}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.ListProducts(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, response)

}

func DeleteProduct(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err1 := strconv.Atoi(idstr)
	if err1 != nil {
		log.Println("Unable to convert to integer")
	}

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductIDRequest{ProductId: uint32(id)}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.DeleteProduct(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, response)

}

func UpdateProduct(ctx *gin.Context) {
	var prod *models.Product

	if err := ctx.BindJSON(&prod); err != nil {
		log.Println("Unable to Bind JSon")
		log.Println(prod)
		ctx.JSON(http.StatusInternalServerError, err)

	}

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.ProductRequest{ProductId: uint32(prod.ProductID), Name: prod.Name, Price: float32(prod.Price), Description: prod.Description}
	client := protobuf.NewProductServiceClient(conn)

	response, err := client.UpdateProduct(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, response)

}
