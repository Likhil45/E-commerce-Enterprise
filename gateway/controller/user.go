package controller

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf"
	"e-commerce/write-db-service/store"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func CreateUser(c *gin.Context) {

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Unable to bind JSON", err)
	}
	conn, err := grpc.Dial(":50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcRequest := &protobuf.RegisterUserRequest{
		Username: user.Username, Userid: int32(user.UserID), Email: user.Email, Password: user.Password,
	}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.RegisterUser(context.Background(), grpcRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	c.JSON(http.StatusOK, response)

}
func Login(ctx *gin.Context) {
	type Usercred struct {
		Username string `json:"user_name"`
		Password string `json:"password"`
	}
	var user Usercred
	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := grpc.Dial(":50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcRequest := &protobuf.AuthenticateUserRequest{
		Username: user.Username, Password: user.Password,
	}
	log.Println(user)
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.AuthenticateUser(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	log.Println(response)

	ctx.JSON(http.StatusOK, gin.H{"token": response.Token})
}

func GetUser(ctx *gin.Context) {

	idstr := ctx.Param("id")
	id, err1 := strconv.Atoi(idstr)
	if err1 != nil {
		log.Println("Unable to convert to integer")
	}
	log.Println(id)
	conn, err := grpc.Dial(":50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcRequest := &protobuf.GetUserRequest{UserId: int32(id)}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.GetUser(context.Background(), grpcRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, response)

}

func GetAllUsers(c *gin.Context) {
	var users []models.User

	// Fetch all users from the database
	fmt.Println("store.DB in controller:", store.DB)

	if store.DB == nil {
		log.Println("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	if err := store.DB.Find(&users).Error; err != nil {
		log.Println("Error fetching users:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Send users as a JSON response
	c.JSON(http.StatusOK, users)
}
