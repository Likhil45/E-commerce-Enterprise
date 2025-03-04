package services

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"

	"e-commerce/write-db-service/store"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm/clause"
)

type OrderService struct {
	protobuf.UnimplementedOrderServiceServer
	KafkaClient protobuf.KafkaProducerServiceClient
	InvClient   protobuf.InventoryServiceClient
}

func NewOrderService(kafkaClient protobuf.KafkaProducerServiceClient) *OrderService {
	return &OrderService{KafkaClient: kafkaClient}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *protobuf.OrderRequest) (*protobuf.OrderResponse, error) {

	order := models.Order{

		UserID: uint32(req.UserId),
	}

	for _, item := range req.Items {

		order.OrderItems = append(order.OrderItems, models.OrderItem{
			ProductID: uint32(item.ProductId),
			Quantity:  uint(item.Quantity),
			Price:     float64(item.Price),
		})
		req.TotalAmount += item.Price
	}
	order.TotalPrice = float64(req.TotalAmount)

	log.Println("Checking order before inserting to DB", order)

	// Insert order with conflict handling
	if err := store.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoNothing: true, // ✅ Ignores duplicate inserts
	}).Create(&order).Error; err != nil {
		log.Println("Unable to create record", err)
		return nil, err
	}
	log.Println(order)

	// //Check order availability
	// for _, val := range req.Items {

	// 	grpcRequestI := &protobuf.StockRequest{ProductId: uint32(val.ProductId)}

	// 	conn1, err2 := grpc.Dial(":50001", grpc.WithInsecure())
	// 	if err2 != nil {
	// 		log.Fatalf("did not connect: %v", err2)
	// 	}
	// 	defer conn1.Close()
	// 	client1 := protobuf.NewInventoryServiceClient(conn1)

	// 	res, err4 := client1.TrackStock(ctx, grpcRequestI)
	// 	if err4 != nil || res.Quantity < uint32(val.Quantity) {
	// 		// c.JSON(http.StatusConflict, gin.H{"error": "Insufficient stock"})
	// 		log.Println("Stock not available", err4)
	// 		break
	// 	}

	// }

	// Publish OrderCreated event to Kafka
	orderJSON, err4 := json.Marshal(req)
	if err4 != nil {
		log.Println("Unable to Marshal json", err4)

	}
	kafkaReq := &protobuf.PublishRequest{
		Topic:     "OrderCreated",
		EventType: "Orders",
		Message:   string(orderJSON),
	}

	// Call Kafka Producer Service via gRPC
	_, err := s.KafkaClient.PublishMessage(ctx, kafkaReq)
	if err != nil {
		log.Printf("Failed to publish message to Kafka Producer Service: %v", err)
		return nil, err
	}

	return &protobuf.OrderResponse{
		OrderId:     req.OrderId,
		UserId:      req.UserId,
		Items:       req.Items,
		TotalAmount: req.TotalAmount,
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *protobuf.OrderIDRequest) (*protobuf.OrderResponse, error) {
	var order models.Order
	if err := store.DB.Preload("Items").Where("id = ?", req.OrderId).First(&order).Error; err != nil {
		return nil, err
	}

	var items []*protobuf.OrderItem
	for _, item := range order.OrderItems {
		items = append(items, &protobuf.OrderItem{
			ProductId: uint32(item.ProductID),
			Quantity:  uint32(item.Quantity),
			Price:     float32(item.Price),
		})
	}

	return &protobuf.OrderResponse{
		OrderId:     uint32(order.OrderID),
		UserId:      uint32(order.UserID),
		Items:       items,
		TotalAmount: float32(order.TotalPrice),
	}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *protobuf.Empty) (*protobuf.OrderListResponse, error) {
	var orders []models.Order
	if err := store.DB.Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}

	var orderResponses []*protobuf.OrderResponse
	for _, order := range orders {
		var items []*protobuf.OrderItem
		for _, item := range order.OrderItems {
			items = append(items, &protobuf.OrderItem{
				ProductId: uint32(item.ProductID),
				Quantity:  uint32(item.Quantity),
				Price:     float32(item.Price),
			})
		}
		orderResponses = append(orderResponses, &protobuf.OrderResponse{
			OrderId:     uint32(order.OrderID),
			UserId:      uint32(order.UserID),
			Items:       items,
			TotalAmount: float32(order.TotalPrice),
		})
	}

	return &protobuf.OrderListResponse{Orders: orderResponses}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *protobuf.OrderRequest) (*protobuf.OrderResponse, error) {
	var order models.Order
	if err := store.DB.Where("id = ?", req.OrderId).First(&order).Error; err != nil {
		return nil, err
	}

	order.TotalPrice = float64(req.TotalAmount)
	store.DB.Save(&order)

	// Publish OrderUpdated event to Kafka
	orderJSON, _ := json.Marshal(req)
	kafkaReq := &protobuf.PublishRequest{
		Topic:     "Orders",
		EventType: "OrderUpdated",
		Message:   string(orderJSON),
	}

	_, err := s.KafkaClient.PublishMessage(ctx, kafkaReq)
	if err != nil {
		log.Printf("Failed to publish update event: %v", err)
		return nil, err
	}

	return &protobuf.OrderResponse{
		OrderId:     req.OrderId,
		UserId:      req.UserId,
		TotalAmount: req.TotalAmount,
	}, nil
}
func (s *OrderService) DeleteOrder(ctx context.Context, req *protobuf.OrderIDRequest) (*protobuf.Empty, error) {
	if err := store.DB.Where("id = ?", req.OrderId).Delete(&models.Order{}).Error; err != nil {
		return nil, err
	}
	// Publish OrderDeleted event to Kafka
	kafkaReq := &protobuf.PublishRequest{
		Topic:     "Orders",
		EventType: "OrderDeleted",
		Message:   fmt.Sprintf("{\"order_id\": %d}", req.OrderId),
	}

	_, err := s.KafkaClient.PublishMessage(ctx, kafkaReq)
	if err != nil {
		log.Printf("Failed to publish delete event: %v", err)
		return nil, err
	}
	return &protobuf.Empty{}, nil
}
