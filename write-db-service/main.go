package main

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf"
	"e-commerce/write-db-service/services"
	"e-commerce/write-db-service/store"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
)

type writeDB struct {
	protobuf.UnimplementedUserServiceServer
}

var JWTSecret = []byte("your_jwt_secret_key")

func (w *writeDB) RegisterUser(ctx context.Context, req *protobuf.RegisterUserRequest) (*protobuf.UserResponse, error) {

	var user models.User
	user.Email = req.GetEmail()
	user.UserID = int(req.GetUserid())
	user.Password = req.GetPassword()
	user.Username = req.GetUsername()

	if err := store.DB.Create(&user).Error; err != nil {
		log.Printf("Error while creating user: %v", err)
		return nil, err
	}

	return &protobuf.UserResponse{UserId: req.GetUserid(), Username: req.GetUsername(), Email: req.GetEmail()}, nil
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
	if err := store.DB.Where("user_id = ?", req.GetUserId()).First(&user).Error; err != nil {
		return nil, err
	}
	return &protobuf.UserResponse{UserId: int32(user.UserID), Username: user.Username, Email: user.Email}, nil
}

func GenerateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(JWTSecret)
}

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
	grpcServer := grpc.NewServer()

	//User
	protobuf.RegisterUserServiceServer(grpcServer, &writeDB{})

	//Product
	protobuf.RegisterProductServiceServer(grpcServer, &services.WriteProduct{})
	err1 := grpcServer.Serve(lsn)
	if err1 != nil {
		log.Println("Unable to Serve ", err1)
	}
	fmt.Println("store.DB in main:", store.DB)

}
