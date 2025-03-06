package payment

import (
	"context"
	"e-commerce/protobuf/protobuf"
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

type PaymentService struct {
	protobuf.UnimplementedPaymentServiceServer
	kafkaClient protobuf.KafkaProducerServiceClient
	userClient  protobuf.UserServiceClient
}

func NewPaymentService(kafkaClient protobuf.KafkaProducerServiceClient, userClient protobuf.UserServiceClient) *PaymentService {
	return &PaymentService{kafkaClient: kafkaClient, userClient: userClient}
}

func (p *PaymentService) ProcessPayment(ctx context.Context, req *protobuf.PaymentRequest) (*protobuf.PaymentResponse, error) {
	log.Printf("Processing payment for OrderID: %d, Amount: %.2f, Method: %s\n", req.OrderId, req.Amount, req.Method)
	userReq := &protobuf.GetUserRequest{UserId: req.UserId}

	user, err1 := p.userClient.GetUserPaymentDetails(ctx, userReq)
	if err1 != nil {
		log.Println("Unable to get User details", err1)
		return nil, err1
	}
	log.Println(user.HasPaymentDetails)
	var paymentStatus string
	if user.HasPaymentDetails {
		paymentStatus = "SUCCESS"
	} else {
		paymentStatus = "FAILURE"
	}
	//Implement payment verification
	paymentID := uuid.New().ID()

	response := &protobuf.PaymentResponse{
		PaymentId: paymentID,
		OrderId:   req.OrderId,
		Amount:    req.Amount,
		Status:    paymentStatus,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Println("Unable to Marshal Payment response")
		return nil, err
	}
	if response.Status == "SUCCESS" {

		p.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "OrderConfirmed", Message: string(responseJSON)})
		return response, nil
	} else {
		p.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "PaymentFailed", Message: string(responseJSON)})
		return response, nil
	}

}
