package main

import (
	"context"
	"e-commerce/logger"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/metrics"
	"e-commerce/write-db-service/services"
	"e-commerce/write-db-service/store"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	logger.InitLogger("write-db-service")
	store.InitDB()
	if store.DB == nil {

		fmt.Println("store.DB in main:", store.DB)
	}
	fmt.Println("store.DB in main:", store.DB)

	lsn, err := net.Listen("tcp", ":50001")
	if err != nil {
		log.Println("Unable to listen to 50001 port", err)
	}

	// Set up gRPC connection to Kafka Producer Service
	kafkaConn, err := grpc.Dial("producer-service:50052", grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor)) // Change to Kafka Producer's actual address
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Kafka Producer Service: %v", err)
	}
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaConn)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor))

	// Register Prometheus metrics for gRPC
	grpc_prometheus.Register(grpcServer)

	//User
	protobuf.RegisterUserServiceServer(grpcServer, &writeDB{})

	//Product
	protobuf.RegisterProductServiceServer(grpcServer, &services.WriteProduct{})

	//Inventory
	protobuf.RegisterDatabaseServiceServer(grpcServer, &services.DatabaseService{})

	//Order
	orderService := &services.OrderService{KafkaClient: kafkaClient}

	protobuf.RegisterOrderServiceServer(grpcServer, orderService)
	go func() {
		err1 := grpcServer.Serve(lsn)
		if err1 != nil {
			logger.Logger.Error("Unable to Serve ", err1)
		}
	}()
	logger.Logger.Info("store.DB in main:", store.DB)

	//metrics
	// Initialize Prometheus metrics
	metrics.InitMetrics()

	// Create a Gin router
	router := gin.Default()

	// Middleware to track metrics
	// router.Use(metrics.MetricsMiddleware())

	// Define your routes
	router.GET("/example", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond) // Simulate processing time
		c.JSON(http.StatusOK, gin.H{"message": "Hello, Prometheus!"})
	})

	// Expose the /metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start the server
	log.Println("Starting write-db-service server on :8001")
	if err := router.Run(":8001"); err != nil {
		logger.Logger.Errorf("Failed to start server: %v", err)
	}

}

type writeDB struct {
	protobuf.UnimplementedUserServiceServer
}

var JWTSecret = []byte("one_piece")

func (w *writeDB) RegisterUser(ctx context.Context, req *protobuf.RegisterUserRequest) (*protobuf.UserResponse, error) {
	// if req.PaymentDetails == nil {
	// 	return nil, fmt.Errorf("PaymentDetails cannot be nil")
	// }
	logger.Logger.Info("Inside the Register user ")
	var user models.User
	user.Email = req.GetEmail()
	user.ID = (req.GetUserId())
	user.Password = req.GetPassword()
	user.Username = req.GetUsername()
	if req.PaymentDetails != nil {
		user.Payment = &models.PaymentDetails{
			ID:            uint(req.PaymentDetails.PaymentId),
			UserID:        (req.GetUserId()),
			PaymentMethod: req.PaymentDetails.PaymentMethod,
			CardNumber:    req.PaymentDetails.CardNumber,
			ExpiryDate:    req.PaymentDetails.ExpiryDate,
		}
	}
	err := store.QueryDatabase("RegisterUser", func() error {
		return store.DB.Create(&user).Error
	})
	if err != nil {
		logger.Logger.Errorf("Error while creating user: %v", err)
		return nil, err
	}

	return &protobuf.UserResponse{UserId: (req.GetUserId()), Username: req.GetUsername(), Email: req.GetEmail()}, nil
}

func (w *writeDB) AuthenticateUser(ctx context.Context, req *protobuf.AuthenticateUserRequest) (*protobuf.AuthResponse, error) {
	var user models.User
	err := store.QueryDatabase("AuthenticateUser", func() error {
		return store.DB.Where("username = ?", req.Username).First(&user).Error
	})
	if err != nil {
		logger.Logger.Errorf("Error while authenticating user: %v", err)
		return nil, err
	}
	if req.Password == user.Password {
		// log.Println(user)

		token, err := GenerateToken(user.ID)
		if err != nil {
			logger.Logger.Error("Unable to Generate Token", err)

		}
		return &protobuf.AuthResponse{Token: token}, nil
	}
	return nil, jwt.ValidationError{}

}

func (w *writeDB) GetUser(ctx context.Context, req *protobuf.GetUserRequest) (*protobuf.UserResponse, error) {
	var user models.User
	err := store.QueryDatabase("GetUser", func() error {
		return store.DB.Where("id = ?", req.GetUserId()).First(&user).Error
	})
	if err != nil {
		return nil, err
	}
	// log.Println("Get User: ", user)
	return &protobuf.UserResponse{UserId: (user.ID), Username: user.Username, Email: user.Email}, nil
}

func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 72).Unix(),
	})
	return token.SignedString(JWTSecret)
}

func (w *writeDB) GetUserPaymentDetails(ctx context.Context, req *protobuf.GetUserRequest) (*protobuf.UserPaymentResponse, error) {
	// user, err := w.GetUser(ctx, &protobuf.GetUserRequest{UserId: req.UserId})
	// if err != nil {
	// 	log.Println("Unable to fetch user details", err)
	// 	return nil, err
	// }
	// pay := user.GetPayment()
	// log.Println(pay)

	var paym models.PaymentDetails
	if err := store.DB.Where("user_id = ?", req.GetUserId()).First(&paym).Error; err != nil {
		logger.Logger.Errorln("unable to find record payment details")
		// return nil, err
	}
	log.Println(paym)

	if paym.CardNumber != "" {
		// log.Println(paym != nil)
		// logger.Logger.Infoln(paym.CardNumber)
		return &protobuf.UserPaymentResponse{HasPaymentDetails: true, Payment: &protobuf.PaymentDetails{PaymentMethod: paym.PaymentMethod, CardNumber: paym.CardNumber, ExpiryDate: paym.ExpiryDate}}, nil
	}
	return &protobuf.UserPaymentResponse{HasPaymentDetails: false, Payment: nil}, nil
}

func (w *writeDB) AddPaymentDetails(ctx context.Context, req *protobuf.AddPaymentDetailsRequest) (*protobuf.AddPaymentDetailsResponse, error) {
	logger.Logger.Infoln("Inside AddPaymentDetails")

	// Validate the request
	if req.GetUserId() == "" || req.GetPaymentDetails() == nil {
		logger.Logger.Errorln("Invalid AddPaymentDetails request: missing required fields")
		return &protobuf.AddPaymentDetailsResponse{Status: "FAILED"}, fmt.Errorf("user_id and payment_details are required")
	}

	// Map the request to the PaymentDetails model
	payment := models.PaymentDetails{
		UserID:        req.GetUserId(),
		PaymentMethod: req.PaymentDetails.GetPaymentMethod(),
		CardNumber:    req.PaymentDetails.GetCardNumber(),
		ExpiryDate:    req.PaymentDetails.GetExpiryDate(),
	}

	// Check if payment details already exist for the user
	var existingPayment models.PaymentDetails
	err := store.QueryDatabase("FindPaymentDetails", func() error {
		return store.DB.Where("user_id = ?", req.GetUserId()).First(&existingPayment).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Errorln("No existing payment details found, creating new record")
			// Create new payment details
			err = store.QueryDatabase("CreatePaymentDetails", func() error {
				return store.DB.Create(&payment).Error
			})
		} else {
			logger.Logger.Errorf("Error while checking existing payment details: %v", err)
			return &protobuf.AddPaymentDetailsResponse{Status: "FAILED"}, err
		}
	} else {
		logger.Logger.Infoln("Updating existing payment details")
		// Update existing payment details
		err = store.QueryDatabase("UpdatePaymentDetails", func() error {
			return store.DB.Model(&existingPayment).Updates(payment).Error
		})
	}

	if err != nil {
		logger.Logger.Errorf("Error while saving payment details: %v", err)
		return &protobuf.AddPaymentDetailsResponse{Status: "FAILED"}, err
	}

	logger.Logger.Infof("Payment details added/updated successfully for UserID=%s", req.GetUserId())
	return &protobuf.AddPaymentDetailsResponse{Status: "SUCCESS"}, nil
}
