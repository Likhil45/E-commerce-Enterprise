package notification

import (
	"context"
	"e-commerce/protobuf/protobuf"
	"log"
)

// type NotificationService struct {
// 	protobuf.UnimplementedNotificationServiceServer
// }

// func (n *NotificationService) SendNotification(ctx context.Context, req *protobuf.NotificationRequest) (*protobuf.NotificationResponse, error) {
// 	conn, err := grpc.Dial(":50010", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	client := protobuf.NewRedisServiceClient(conn)

// 	usrID := strconv.Itoa(int(req.UserId))

// 	grpcRequest := &protobuf.GetRequest{Key: usrID}

// 	response, err := client.GetData(ctx, grpcRequest)
// 	if err != nil {
// 		log.Println("grpc error")
// 		return nil, err
// 	}
// 	grpcResponse := &protobuf.NotificationResponse{Status: response.Value}
// 	log.Println("Response of setting notification to redis", response)
// 	return (*protobuf.NotificationResponse)(grpcResponse), nil
// }

type NotificationService struct {
	protobuf.UnimplementedNotificationServiceServer
	redisClient protobuf.RedisServiceClient
}

// NewNotificationService initializes the service
func NewNotificationService(redisClient protobuf.RedisServiceClient) *NotificationService {
	return &NotificationService{
		redisClient: redisClient,
	}
}

func (n *NotificationService) SendNotification(ctx context.Context, req *protobuf.NotificationRequest) (*protobuf.NotificationResponse, error) {
	// usrID := strconv.Itoa(int(req.UserId))
	grpcRequest := &protobuf.GetRequest{Key: req.UserId}

	response, err := n.redisClient.GetData(ctx, grpcRequest)
	if err != nil {
		log.Println("gRPC error:", err)
		return nil, err
	}

	log.Println("Response of setting notification ", response)
	return &protobuf.NotificationResponse{Status: response.Value}, nil
}
