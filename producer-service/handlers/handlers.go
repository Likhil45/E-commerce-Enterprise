package handlers

import (
	"context"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// KafkaProducerService handles gRPC requests and sends messages to Kafka
type KafkaProducerService struct {
	protobuf.UnimplementedKafkaProducerServiceServer
	producer sarama.SyncProducer
	// topic    string
}

// NewKafkaProducer initializes the Kafka producer
func NewKafkaProducer(brokers []string) (*KafkaProducerService, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Failed to start Kafka producer: %v", err)
		return nil, err
	}

	return &KafkaProducerService{producer: producer}, nil
}

// PublishMessage receives gRPC requests and sends events to Kafka
func (p *KafkaProducerService) PublishMessage(ctx context.Context, req *protobuf.PublishRequest) (*protobuf.PublishResponse, error) {

	//Validate Topic name
	// if req.Topic == "" {
	// 	err := fmt.Errorf("invalid Kafka topic: topic name is empty")
	// 	log.Println("❌", err)
	// 	return nil, err
	// }

	if req.EventType == "" {
		err := fmt.Errorf("invalid Kafka event type: event type is empty")
		log.Println("❌", err)
		return nil, err
	}

	log.Println(req.EventType)

	msg := &sarama.ProducerMessage{
		Topic: req.Topic,
		Key:   sarama.StringEncoder(req.EventType),
		Value: sarama.StringEncoder(req.Message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		msg.Topic,
		partition,
		offset)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return nil, err
	}
	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		msg.Topic,
		partition,
		offset)
	log.Printf("Published event: %s", req.EventType)

	return &protobuf.PublishResponse{Status: "Message sent to Kafka"}, nil
}
