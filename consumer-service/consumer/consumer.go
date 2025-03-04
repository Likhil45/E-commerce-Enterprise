package consumer

import (
	"context"
	conshand "e-commerce/consumer-service/handler"
	"e-commerce/protobuf/protobuf"
	"encoding/json"
	"log"

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
		log.Printf("Failed to start Kafka consumer: %v", err)
		return nil, err
	}

	return &KafkaConsumerService{consumer: consumer, topics: topics, groupID: groupID, handler: ConsumerGroupHandler{}}, nil
}

// ProcessMessage processes Kafka messages
// ProcessMessage listens for messages from Kafka and processes them
func (c *KafkaConsumerService) ProcessMessage(ctx context.Context) {
	log.Println("Starting Kafka Consumer...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer stopped")
			return
		default:
			err := c.consumer.Consume(ctx, c.topics, &c.handler)
			if err != nil {
				log.Printf("Error consuming Kafka message: %v", err)
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
	log.Println("Inside ConsumeClaim")

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Println("Message channel closed, stopping processing")
				return nil
			}

			log.Printf("Received message: Topic=%s, Partition=%d, Offset=%d, Message=%s\n",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Value))

			// Handle event based on topic
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
			default:
				log.Printf("Unknown topic: %s", msg.Topic)
			}

			// Mark message as processed
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			log.Println("Consumer session canceled")
			return nil
		}
	}
}

// Order Service Handler
func ProcessOrderCreated(message string) {

	log.Println(message)
	log.Println("Received OrderCreated event:", message)

	//  Parse the incoming message
	var order protobuf.OrderResponse
	err := json.Unmarshal([]byte(message), &order)
	if err != nil {
		log.Println("Error parsing OrderCreated event:", err)
		return
	}

	//  Validate event data
	if order.OrderId == 0 || order.UserId == 0 || len(order.Items) == 0 {
		log.Println("Invalid OrderCreated event: missing required fields")

		return
	}
	// var error1 int32
	// for _, val := range order.Items {

	// 	inventoryReq := &protobuf.StockUpdateRequest{ProductId: val.ProductId, Quantity: val.Quantity}
	// 	inventoryResp, err := conshand.CallInventoryService(inventoryReq)
	// 	if err != nil {
	// 		log.Println("Inventory processing failed:", err)
	// 	}
	// 	if inventoryResp.Status == "Insufficient Stock" || inventoryResp.Status == "Not Found" {
	// 		error1++
	// 	}

	// }

	// if error1 == 0 {

	paymentReq := &protobuf.PaymentRequest{OrderId: order.OrderId, Amount: order.TotalAmount, Method: "credit_card", UserId: order.UserId}

	paymentResponse, err := conshand.CallPaymentService(paymentReq)
	if err != nil {
		log.Println("Payment processing failed:", err)
		return
	}
	notReq := &protobuf.NotificationRequest{UserId: order.GetUserId(), Message: "Your Order is Created"}

	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		log.Println("Unable to set Nofication", err)
		return
	}
	log.Printf("\nPayment Process Response: %v", paymentResponse)
	log.Printf("\nNotification Response: %v", notResp)

	// }
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
	log.Println("Processing PaymentProcessed event:", message)
	var payment protobuf.PaymentResponse
	err := json.Unmarshal([]byte(message), &payment)
	if err != nil {
		log.Println("unable to Unmarshall the data from paymentprocess event")
	}

	//Validate with payment status
	if payment.Status != "SUCCESS" {
		log.Println("Payment Failed")
	}
	//Order
	notReq := &protobuf.NotificationRequest{UserId: payment.UserId, Message: "Stock is available"}

	res, err := conshand.CallNotificationService(notReq)
	if err != nil {
		log.Println("Failed to call Notifaction Service", err)
	}
	log.Printf("\n Notification Response: %v", res)
}

// Notification Service Handler
func ProcessOrderConfirmed(message string) {
	log.Println("Processing OrderConfirmed event:", message)
	var payresp protobuf.PaymentResponse

	err := json.Unmarshal([]byte(message), &payresp)
	if err != nil {
		log.Println("Unable to Unamrshal Payment Request")
		return
	}
	notReq := &protobuf.NotificationRequest{UserId: payresp.GetUserId(), Message: "Your Order is Confirmed"}
	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		log.Println("Unable to set Nofication", err)
		return
	}
	log.Printf("\nNotification Response: %v", notResp)

}

func ProductOutOfStock(message string) {
	log.Println("Processing Out of Stock event:", message)
	var prodId uint32

	err := json.Unmarshal([]byte(message), &prodId)
	if err != nil {
		log.Println("Unable to Unamrshal OutofStock Request")
		return
	}

	//Need to Implement the notification properly
	notReq := &protobuf.NotificationRequest{UserId: prodId, Message: "Your Order is Confirmed"}
	notResp, err2 := conshand.CallNotificationService(notReq)
	if err2 != nil {
		log.Println("Unable to set Nofication", err)
		return
	}
	log.Printf("\nNotification Response: %v", notResp)

}
