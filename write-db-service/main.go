package main

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/services"
	"e-commerce/write-db-service/store"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
)

func main() {

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
	kafkaConn, err := grpc.Dial("producer-service:50052", grpc.WithInsecure()) // Change to Kafka Producer's actual address
	if err != nil {
		log.Fatalf("Failed to connect to Kafka Producer Service: %v", err)
	}
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaConn)

	grpcServer := grpc.NewServer()

	//User
	protobuf.RegisterUserServiceServer(grpcServer, &writeDB{})

	//Product
	protobuf.RegisterProductServiceServer(grpcServer, &services.WriteProduct{})

	//Inventory
	protobuf.RegisterDatabaseServiceServer(grpcServer, &services.DatabaseService{})

	//Order
	orderService := &services.OrderService{KafkaClient: kafkaClient}

	protobuf.RegisterOrderServiceServer(grpcServer, orderService)

	err1 := grpcServer.Serve(lsn)
	if err1 != nil {
		log.Println("Unable to Serve ", err1)
	}
	fmt.Println("store.DB in main:", store.DB)

}

type writeDB struct {
	protobuf.UnimplementedUserServiceServer
}

var JWTSecret = []byte("your_jwt_secret_key")

func (w *writeDB) RegisterUser(ctx context.Context, req *protobuf.RegisterUserRequest) (*protobuf.UserResponse, error) {
	// if req.PaymentDetails == nil {
	// 	return nil, fmt.Errorf("PaymentDetails cannot be nil")
	// }
	log.Println("Inside the Register user ")
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

	if err := store.DB.Create(&user).Error; err != nil {
		log.Printf("Error while creating user: %v", err)
		return nil, err
	}

	return &protobuf.UserResponse{UserId: (req.GetUserId()), Username: req.GetUsername(), Email: req.GetEmail()}, nil
}

func (w *writeDB) AuthenticateUser(ctx context.Context, req *protobuf.AuthenticateUserRequest) (*protobuf.AuthResponse, error) {
	var user models.User
	if err := store.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		// log.Println(user)

		return nil, err
	}
	if req.Password == user.Password {
		log.Println(user)

		token, err := GenerateToken(user.ID)
		if err != nil {
			log.Println("Unable to Generate Token", err)

		}
		return &protobuf.AuthResponse{Token: token}, nil
	}
	return nil, jwt.ValidationError{}

}

func (w *writeDB) GetUser(ctx context.Context, req *protobuf.GetUserRequest) (*protobuf.UserResponse, error) {
	var user models.User
	if err := store.DB.Where("id = ?", req.GetUserId()).First(&user).Error; err != nil {
		return nil, err
	}
	log.Println("Get User: ", user)
	return &protobuf.UserResponse{UserId: (user.ID), Username: user.Username, Email: user.Email}, nil
}

func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
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
		log.Println("unable to find record payment details")
		// return nil, err
	}
	log.Println(paym)

	if paym.CardNumber != "" {
		// log.Println(paym != nil)
		return &protobuf.UserPaymentResponse{HasPaymentDetails: true, Payment: &protobuf.PaymentDetails{PaymentMethod: paym.PaymentMethod, CardNumber: paym.CardNumber, ExpiryDate: paym.ExpiryDate}}, nil
	}
	return &protobuf.UserPaymentResponse{HasPaymentDetails: false, Payment: nil}, nil
}
