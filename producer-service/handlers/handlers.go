package handlers

import (
	"context"
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	KafkaMessagesPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_published_total",
			Help: "Total number of Kafka messages published",
		},
		[]string{"topic", "status"},
	)

	GrpcCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_calls_total",
			Help: "Total number of gRPC calls made to the producer-service",
		},
		[]string{"method", "status"},
	)
)

func Init() {
	prometheus.MustRegister(KafkaMessagesPublished)
	prometheus.MustRegister(GrpcCallsTotal)
}

// KafkaProducerService handles gRPC requests and sends messages to Kafka
type KafkaProducerService struct {
	protobuf.UnimplementedKafkaProducerServiceServer
	producer sarama.SyncProducer
}

// NewKafkaProducer initializes the Kafka producer
func NewKafkaProducer(brokers []string) (*KafkaProducerService, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logger.Logger.Errorf("Failed to start Kafka producer: %v", err)
		return nil, err
	}

	logger.Logger.Info("Kafka producer initialized successfully")
	return &KafkaProducerService{producer: producer}, nil
}

// PublishMessage receives gRPC requests and sends events to Kafka
func (p *KafkaProducerService) PublishMessage(ctx context.Context, req *protobuf.PublishRequest) (*protobuf.PublishResponse, error) {
	logger.Logger.Infof("Received PublishMessage request: Topic=%s, EventType=%s", req.Topic, req.EventType)

	// Validate EventType
	if req.EventType == "" {
		err := fmt.Errorf("invalid Kafka event type: event type is empty")
		logger.Logger.Errorf("Validation failed: %v", err)
		GrpcCallsTotal.WithLabelValues("PublishMessage", "failed").Inc()
		return nil, err
	}

	msg := &sarama.ProducerMessage{
		Topic: req.Topic,
		Key:   sarama.StringEncoder(req.EventType),
		Value: sarama.StringEncoder(req.Message),
	}
	logger.Logger.Infof("Preparing to send message to Kafka: Topic=%s, EventType=%s", req.Topic, req.EventType)

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		logger.Logger.Errorf("Failed to publish message to Kafka: Topic=%s, EventType=%s, Error=%v", req.Topic, req.EventType, err)
		KafkaMessagesPublished.WithLabelValues(req.Topic, "failed").Inc()
		GrpcCallsTotal.WithLabelValues("PublishMessage", "failed").Inc()
		return nil, err
	}

	logger.Logger.Infof("Message published to Kafka successfully: Topic=%s, Partition=%d, Offset=%d", msg.Topic, partition, offset)
	KafkaMessagesPublished.WithLabelValues(req.Topic, "success").Inc()
	GrpcCallsTotal.WithLabelValues("PublishMessage", "success").Inc()

	return &protobuf.PublishResponse{Status: "Message sent to Kafka"}, nil
}
