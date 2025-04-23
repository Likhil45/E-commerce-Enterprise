package controller

import (
	"context"
	"e-commerce/logger"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/store"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		logger.Logger.Errorf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userID := uuid.New().String()
	user.ID = userID
	user.Payment.UserID = user.ID
	logger.Logger.Infof("Creating user: %+v", user)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Errorf("Failed to connect to gRPC service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gRPC service"})
		return
	}
	defer conn.Close()

	grpcRequest := &protobuf.RegisterUserRequest{
		Username: user.Username,
		UserId:   user.ID,
		Email:    user.Email,
		Password: user.Password,
		PaymentDetails: &protobuf.PaymentDetails{
			PaymentMethod: user.Payment.PaymentMethod,
			CardNumber:    user.Payment.CardNumber,
			ExpiryDate:    user.Payment.ExpiryDate,
		},
	}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.RegisterUser(c, grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error in user signup: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error in user signup: %v", err)})
		return
	}

	logger.Logger.Infof("User created successfully: %+v", response)
	c.JSON(http.StatusOK, response)
}

func Login(ctx *gin.Context) {
	type Usercred struct {
		Username string `json:"user_name"`
		Password string `json:"password"`
	}
	var user Usercred
	if err := ctx.BindJSON(&user); err != nil {
		logger.Logger.Errorf("Invalid login payload: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Errorf("Failed to connect to gRPC service: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gRPC service"})
		return
	}
	defer conn.Close()

	grpcRequest := &protobuf.AuthenticateUserRequest{
		Username: user.Username,
		Password: user.Password,
	}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.AuthenticateUser(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error during login: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("User logged in successfully: %s", user.Username)
	ctx.JSON(http.StatusOK, gin.H{"token": response.Token})
}

func GetUser(ctx *gin.Context) {
	idstr := ctx.Param("id")
	logger.Logger.Infof("Fetching user with ID: %s", idstr)

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Errorf("Failed to connect to gRPC service: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gRPC service"})
		return
	}
	defer conn.Close()

	grpcRequest := &protobuf.GetUserRequest{UserId: idstr}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.GetUser(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while fetching user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("User fetched successfully: %+v", response)
	ctx.JSON(http.StatusOK, response)
}

func GetAllUsers(c *gin.Context) {
	logger.Logger.Info("Fetching all users from the database")

	var users []models.User
	if store.DB == nil {
		logger.Logger.Error("Database connection is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	if err := store.DB.Find(&users).Error; err != nil {
		logger.Logger.Errorf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	logger.Logger.Infof("Fetched %d users successfully", len(users))
	c.JSON(http.StatusOK, users)
}

func AddPaymentDetails(ctx *gin.Context) {
	var payment models.PaymentDetails
	if err := ctx.BindJSON(&payment); err != nil {
		logger.Logger.Errorf("Invalid payment details payload: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Errorf("Failed to connect to gRPC service: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gRPC service"})
		return
	}
	defer conn.Close()

	grpcRequest := &protobuf.AddPaymentDetailsRequest{
		UserId: payment.UserID,
		PaymentDetails: &protobuf.PaymentDetails{
			PaymentMethod: payment.PaymentMethod,
			CardNumber:    payment.CardNumber,
			ExpiryDate:    payment.ExpiryDate,
		},
	}
	client := protobuf.NewUserServiceClient(conn)

	response, err := client.AddPaymentDetails(context.Background(), grpcRequest)
	if err != nil {
		logger.Logger.Errorf("gRPC error while adding payment details: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	logger.Logger.Infof("Payment details added successfully for UserID=%s", payment.UserID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Payment details added successfully", "status": response.Status})
}
