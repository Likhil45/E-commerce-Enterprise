package controller

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/store"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func CreateUser(c *gin.Context) {

	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	userID := uuid.New().String()
	user.ID = userID
	user.Payment.UserID = user.ID
	log.Println(user)
	if user.Payment != nil {
		log.Printf("Payment UserID: %s\n", user.Payment.UserID)
		log.Printf("Payment method: %s\n", user.Payment.PaymentMethod)
		log.Printf("Card number: %s\n", user.Payment.CardNumber)
		log.Printf("Expiry date: %s\n", user.Payment.ExpiryDate)
		log.Printf("CVV: %s\n", user.Payment.CVV)

		// You can now process the payment details further if needed
	}
	conn, err := grpc.Dial(":50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcRequest := &protobuf.RegisterUserRequest{
		Username: user.Username, UserId: (user.ID), Email: user.Email, Password: user.Password, PaymentDetails: &protobuf.PaymentDetails{PaymentMethod: user.Payment.PaymentMethod, CardNumber: user.Payment.CardNumber, ExpiryDate: user.Payment.ExpiryDate},
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

	grpcRequest := &protobuf.GetUserRequest{UserId: (idstr)}
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

func AddPaymentDetails(c *gin.Context) {

}
