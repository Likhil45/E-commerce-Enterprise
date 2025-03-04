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
}

func NewPaymentService(kafkaClient protobuf.KafkaProducerServiceClient) *PaymentService {
	return &PaymentService{kafkaClient: kafkaClient}
}

func (p *PaymentService) ProcessPayment(ctx context.Context, req *protobuf.PaymentRequest) (*protobuf.PaymentResponse, error) {
	log.Printf("Processing payment for OrderID: %d, Amount: %.2f, Method: %s\n", req.OrderId, req.Amount, req.Method)
	paymentID := uuid.New().ID()
	paymentStatus := "SUCCESS"

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
	p.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "OrderConfirmed", Message: string(responseJSON)})
	return response, nil
}
