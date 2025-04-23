package consumer

import (
	"context"
	"e-commerce/consumer-service/consmetrics"
	conshand "e-commerce/consumer-service/handler"
	"e-commerce/logger"
	"strconv"

	"e-commerce/protobuf/protobuf"
	"encoding/json"

	"github.com/IBM/sarama"
)

// KafkaConsumerService listens for events and processes them
type KafkaConsumerService struct {
	consumer sarama.ConsumerGroup
	topics   []string
	groupID  string
	handler  ConsumerGroupHandler
}

// NewKafkaConsumer initializes the Kafka consumer
func NewKafkaConsumer(brokers []string, groupID string, topics []string) (*KafkaConsumerService, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		logger.Logger.Errorf("Failed to start Kafka consumer: %v", err)
		return nil, err
	}
	logger.Logger.Info("Kafka consumer initialized successfully")
	return &KafkaConsumerService{consumer: consumer, topics: topics, groupID: groupID, handler: ConsumerGroupHandler{}}, nil
}

// ProcessMessage listens for messages from Kafka and processes them
func (c *KafkaConsumerService) ProcessMessage(ctx context.Context) {
	logger.Logger.Info("Starting Kafka Consumer...")

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Kafka consumer stopped")
			return
		default:
			err := c.consumer.Consume(ctx, c.topics, &c.handler)
			if err != nil {
				logger.Logger.Errorf("Error consuming Kafka message: %v", err)
			}
		}
	}
}

// ConsumerGroupHandler handles Kafka messages
type ConsumerGroupHandler struct{}

// Setup runs at the beginning of a new session
func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup runs at the end of a session
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes each Kafka message
func (h *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Logger.Info("Inside ConsumeClaim")

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				logger.Logger.Info("Message channel closed, stopping processing")
				return nil
			}

			logger.Logger.Infof("Received message: Topic=%s, Partition=%d, Offset=%d, Message=%s",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Value))

			// Handle event based on topic
			status := "success"
			switch msg.Topic {
			case "OrderCreated":
				ProcessOrderCreated(string(msg.Value))
			case "InventoryReserved":
				ProcessInventoryReserved(string(msg.Value))
			case "PaymentProcessed":
				ProcessPaymentProcessed(string(msg.Value))
			case "OrderConfirmed":
				ProcessOrderConfirmed(string(msg.Value))
			case "OutOfStock":
				ProductOutOfStock(string(msg.Value))
			case "PaymentFailed":
				PaymentFailed(string(msg.Value))
			default:
				logger.Logger.Warnf("Unknown topic: %s", msg.Topic)
				status = "unknown"
			}

			// Increment Kafka message processed counter
			consmetrics.KafkaMessagesProcessed.WithLabelValues(msg.Topic, status).Inc()

			// Mark message as processed
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			logger.Logger.Info("Consumer session canceled")
			return nil
		}
	}
}

// Order Service Handler
func ProcessOrderCreated(message string) {
	logger.Logger.Infof("Received OrderCreated event: %s", message)

	// Parse the incoming message
	var order protobuf.OrderResponse
	err := json.Unmarshal([]byte(message), &order)
	if err != nil {
		logger.Logger.Errorf("Error parsing OrderCreated event: %v", err)
		return
	}

	// Validate event data
	if order.OrderId == 0 || order.UserId == "" || len(order.Items) == 0 {
		logger.Logger.Warn("Invalid OrderCreated event: missing required fields")
		return
	}

	paymentReq := &protobuf.PaymentRequest{OrderId: order.OrderId, Amount: order.TotalAmount, Method: "credit_card", UserId: order.UserId}

	paymentResponse, err := conshand.CallPaymentService(paymentReq)
	if err != nil {
		logger.Logger.Errorf("Payment processing failed: %v", err)
		return
	}
	logger.Logger.Infof("Payment Process Response: %+v", paymentResponse)
}

// Inventory Service Handler

// Not Implementing
func ProcessInventoryReserved(message string) {
	// // Need to implement order history
	// log.Println("Processing InventoryReserved event:", message)
	// var prodId uint32
	// err := json.Unmarshal([]byte(message), &prodId)
	// if err != nil {
	// 	log.Println("unable to Unmarshall the data from Inventory Reserved event")
	// }

	// paymentReq := &protobuf.PaymentRequest{OrderId: order.OrderId, Amount: order.TotalAmount, Method: "credit_card", UserId: order.UserId}

	// paymentResponse, err := conshand.CallPaymentService(paymentReq)
	// if err != nil {
	// 	log.Println("Payment processing failed:", err)
	// 	return
	// }
	// res, err := conshand.CallNotificationService(notReq)

}

// Payment Service Handler

// Not implementing
func ProcessPaymentProcessed(message string) {
	logger.Logger.Infoln("Processing PaymentProcessed event:", message)
	var payment protobuf.PaymentResponse
	err := json.Unmarshal([]byte(message), &payment)
	if err != nil {
		logger.Logger.Error("unable to Unmarshall the data from paymentprocess event")
	}

	//Validate with payment status
	if payment.Status != "SUCCESS" {
		logger.Logger.Infoln("Payment Failed")
	}
	//Order
	notReq := &protobuf.NotificationRequest{UserId: payment.UserId, Message: "Stock is available"}

	res, err := conshand.CallNotificationService(notReq)
	if err != nil {
		logger.Logger.Error("Failed to call Notifaction Service", err)
	}
	logger.Logger.Infof("\n Notification Response: %v", res)
}

// Notification Service Handler
func ProcessOrderConfirmed(message string) {
	logger.Logger.Infoln("Processing OrderConfirmed event:", message)
	var payresp protobuf.PaymentResponse

	err := json.Unmarshal([]byte(message), &payresp)
	if err != nil {
		logger.Logger.Error("Unable to Unamrshal Payment Request")
		return
	}

	notReq := &protobuf.NotificationRequest{UserId: payresp.GetUserId(), Message: "Your Order is Confirmed", OrderId: strconv.Itoa(int(payresp.GetOrderId())), PaymentId: strconv.Itoa(int(payresp.PaymentId)), Total: strconv.Itoa(int(payresp.Amount))}
	logger.Logger.Infof("\nNotification Request: %v", notReq)
	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		logger.Logger.Error("Unable to set Nofication", err)
		return
	}
	logger.Logger.Infof("\nNotification Response: %v", notResp)

}

func ProductOutOfStock(message string) {
	logger.Logger.Infoln("Processing Out of Stock event:", message)
	var prodId uint32

	err := json.Unmarshal([]byte(message), &prodId)
	if err != nil {
		logger.Logger.Error("Unable to Unamrshal OutofStock Request")
		return
	}

	//Need to Implement the notification properly
	notReq := &protobuf.NotificationRequest{UserId: "122", Message: "Your Order is Confirmed"}
	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		logger.Logger.Error("Unable to set Nofication", err)
		return
	}
	logger.Logger.Infof("\nNotification Response: %v", notResp)

}
func PaymentFailed(message string) {
	logger.Logger.Infoln("Processing Out of Stock event:", message)
	var payresp protobuf.PaymentResponse

	err := json.Unmarshal([]byte(message), &payresp)
	if err != nil {
		logger.Logger.Infoln("Unable to Unmarshal Payment Failed event")
		return
	}
	notReq := &protobuf.NotificationRequest{UserId: payresp.GetUserId(), Message: "Your Payment Failed!! Please update your payment details here: localhost:8080/update/pd and Try again!!!", OrderId: strconv.Itoa(int(payresp.GetOrderId())), PaymentId: strconv.Itoa(int(payresp.PaymentId)), Total: strconv.Itoa(int(payresp.Amount))}
	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		logger.Logger.Infoln("Unable to set Nofication", err)
		return
	}
	logger.Logger.Infof("\nNotification Response: %v", notResp)

}
