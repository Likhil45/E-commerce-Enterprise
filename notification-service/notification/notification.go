package notification

import (
	"context"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"log"
	"net/smtp"

	"github.com/prometheus/client_golang/prometheus"
)

var notificationsSentTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notifications_sent_total",
		Help: "Total number of notifications sent",
	},
	[]string{"status"},
)

func Init() {
	prometheus.MustRegister(notificationsSentTotal)
}

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
	log.Printf("SendNotification called with UserId=%s", req.UserId)
	grpcRequest := &protobuf.GetRequest{Key: req.UserId}

	response, err := n.redisClient.GetData(ctx, grpcRequest)
	log.Printf("Sending GetData request to redis-service: Key=%s", grpcRequest.Key)
	if err != nil {
		log.Printf("Failed to get data from redis-service: %v", err)
		notificationsSentTotal.WithLabelValues("failed").Inc() // Increment failure counter
		return nil, err
	}
	SendEmail(req.UserId, req.Total, req.OrderId, req.Message, req.PaymentId)
	log.Println("Response of setting notification ", response)
	notificationsSentTotal.WithLabelValues("success").Inc() // Increment success counter
	return &protobuf.NotificationResponse{Status: response.Value}, nil
}
func SendEmail(usrId string, total string, ordid string, msg string, payId string) {
	from := "likhilcharan.vellanki99@gmail.com"
	password := "mdwb wyri mgkh nhmo" // Use App Password, not regular password

	to := []string{"chill2vell@gmail.com"}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	message := []byte(fmt.Sprintf(
		"Subject: Your Order Confirmation\n\n"+
			"Hello,%s\n\n"+
			"Thank you for your order!\n\n"+
			"Order ID: %s\n"+
			"Payment ID: %s\n"+
			"Total Amount: %s\n\n"+
			"Order Status: %s\n\n"+
			"We appreciate your business and hope to serve you again soon.\n\n"+
			"Best regards,\n"+
			"E-Commerce Platform Team", usrId,
		ordid, payId, total, msg,
	))
	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email sent successfully!")
}
