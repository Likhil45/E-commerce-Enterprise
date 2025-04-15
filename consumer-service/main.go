package main

import (
	"context"
	"e-commerce/consumer-service/consumer"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	brokers := []string{"kafka:9092"}
	groupID := "ecommerce-consumer-group"
	topics := []string{"OrderCreated", "OutOfStock", "InventoryReserved", "PaymentProcessed", "OrderConfirmed", "PaymentFailed"}

	// Initialize the Kafka consumer
	consumerService, err := consumer.NewKafkaConsumer(brokers, groupID, topics)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down consumer...")
		cancel()
	}()

	// Start the consumer
	consumerService.ProcessMessage(ctx)
}
