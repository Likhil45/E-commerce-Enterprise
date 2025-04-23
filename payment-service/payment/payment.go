package payment

import (
	"context"
	"e-commerce/logger"
	metricsprom "e-commerce/payment-service/metrics"
	"e-commerce/protobuf/protobuf"
	"encoding/json"

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
	logger.Logger.Infof("Processing payment for OrderID: %d, Amount: %.2f, Method: %s", req.OrderId, req.Amount, req.Method)

	// Fetch user payment details
	userReq := &protobuf.GetUserRequest{UserId: req.UserId}
	user, err := p.userClient.GetUserPaymentDetails(ctx, userReq)
	if err != nil {
		logger.Logger.Errorf("Unable to get user payment details: %v", err)
		return nil, err
	}
	logger.Logger.Infof("User payment details fetched: HasPaymentDetails=%v", user.HasPaymentDetails)

	// Determine payment status
	var paymentStatus string
	if user.HasPaymentDetails {
		paymentStatus = "SUCCESS"
		metricsprom.PaymentSuccessTotal.Inc()
	} else {
		paymentStatus = "FAILURE"
		metricsprom.PaymentFailureTotal.Inc()
	}

	// Generate payment response
	paymentID := uuid.New().ID()
	response := &protobuf.PaymentResponse{
		PaymentId: paymentID,
		OrderId:   req.OrderId,
		Amount:    req.Amount,
		Status:    paymentStatus,
		UserId:    req.UserId,
	}

	// Marshal response to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Logger.Errorf("Unable to marshal payment response: %v", err)
		return nil, err
	}

	// Publish event to Kafka based on payment status
	if response.Status == "SUCCESS" {
		logger.Logger.Infof("Publishing OrderConfirmed event to Kafka: %s", string(responseJSON))
		p.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{
			Topic:     "OrderConfirmed",
			EventType: "Orders",
			Message:   string(responseJSON),
		})
	} else {
		logger.Logger.Infof("Publishing PaymentFailed event to Kafka: %s", string(responseJSON))
		p.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{
			Topic:     "PaymentFailed",
			EventType: "Orders",
			Message:   string(responseJSON),
		})
	}

	logger.Logger.Infof("Payment processed successfully: %+v", response)
	return response, nil
}
