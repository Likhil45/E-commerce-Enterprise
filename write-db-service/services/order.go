package services

import (
	"context"
	"e-commerce/logger"
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
	logger.Logger.Infof("Received CreateOrder request: %+v", req)

	order := models.Order{
		UserID:  req.UserId,
		OrderID: req.OrderId,
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

	logger.Logger.Infof("Preparing to insert order into DB: %+v", order)
	log.Printf("\nPreparing to insert order into DB: %+v", order)

	// Insert order with conflict handling
	if err := store.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoNothing: true, // Ignores duplicate inserts
	}).Create(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to create order in DB: %v", err)
		return nil, err
	}
	logger.Logger.Infof("Order created successfully in DB: %+v", order)

	// Publish OrderCreated event to Kafka
	orderJSON, err := json.Marshal(req)
	if err != nil {
		logger.Logger.Errorf("Failed to marshal order to JSON: %v", err)
		return nil, err
	}

	kafkaReq := &protobuf.PublishRequest{
		Topic:     "OrderCreated",
		EventType: "Orders",
		Message:   string(orderJSON),
	}

	// Call Kafka Producer Service via gRPC
	_, err = s.KafkaClient.PublishMessage(ctx, kafkaReq)
	if err != nil {
		logger.Logger.Errorf("Failed to publish message to Kafka Producer Service: %v", err)
		return nil, err
	}
	logger.Logger.Infof("OrderCreated event published to Kafka successfully: %+v", kafkaReq)

	return &protobuf.OrderResponse{
		OrderId:     req.OrderId,
		UserId:      req.UserId,
		Items:       req.Items,
		TotalAmount: req.TotalAmount,
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *protobuf.OrderIDRequest) (*protobuf.OrderResponse, error) {
	logger.Logger.Infof("Received GetOrder request from : UserID=%s", req.UserId)

	var order models.Order
	if err := store.DB.Preload("Items").Where("order_id = ?", req.OrderId).First(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to fetch order from DB: OrderID=%d, Error=%v", req.OrderId, err)
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

	logger.Logger.Infof("Order fetched successfully: %+v", order)
	return &protobuf.OrderResponse{
		OrderId:     uint32(order.OrderID),
		UserId:      order.UserID,
		Items:       items,
		TotalAmount: float32(order.TotalPrice),
	}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *protobuf.Empty) (*protobuf.OrderListResponse, error) {
	logger.Logger.Info("Received ListOrders request")

	var orders []models.Order
	if err := store.DB.Preload("Items").Find(&orders).Error; err != nil {
		logger.Logger.Errorf("Failed to fetch orders from DB: %v", err)
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
			UserId:      order.UserID,
			Items:       items,
			TotalAmount: float32(order.TotalPrice),
		})
	}

	logger.Logger.Infof("Orders fetched successfully: TotalOrders=%d", len(orderResponses))
	return &protobuf.OrderListResponse{Orders: orderResponses}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *protobuf.OrderRequest) (*protobuf.OrderResponse, error) {
	logger.Logger.Infof("Received UpdateOrder request: %+v", req)

	var order models.Order
	if err := store.DB.Where("order_id = ?", req.OrderId).First(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to fetch order for update: OrderID=%d, Error=%v", req.OrderId, err)
		return nil, err
	}

	order.TotalPrice = float64(req.TotalAmount)
	if err := store.DB.Save(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to update order in DB: OrderID=%d, Error=%v", req.OrderId, err)
		return nil, err
	}

	// Publish OrderUpdated event to Kafka
	orderJSON, _ := json.Marshal(req)
	kafkaReq := &protobuf.PublishRequest{
		Topic:     "Orders",
		EventType: "OrderUpdated",
		Message:   string(orderJSON),
	}

	_, err := s.KafkaClient.PublishMessage(ctx, kafkaReq)
	if err != nil {
		logger.Logger.Errorf("Failed to publish OrderUpdated event: %v", err)
		return nil, err
	}
	logger.Logger.Infof("OrderUpdated event published successfully: %+v", kafkaReq)

	return &protobuf.OrderResponse{
		OrderId:     req.OrderId,
		UserId:      req.UserId,
		TotalAmount: req.TotalAmount,
	}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *protobuf.OrderIDRequest) (*protobuf.Empty, error) {
	logger.Logger.Infof("Received DeleteOrder request: OrderID=%d", req.OrderId)

	if err := store.DB.Where("id = ?", req.OrderId).Delete(&models.Order{}).Error; err != nil {
		logger.Logger.Errorf("Failed to delete order from DB: OrderID=%d, Error=%v", req.OrderId, err)
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
		logger.Logger.Errorf("Failed to publish OrderDeleted event: %v", err)
		return nil, err
	}
	logger.Logger.Infof("OrderDeleted event published successfully: %+v", kafkaReq)

	return &protobuf.Empty{}, nil
}
func (s *OrderService) GetOrdersByUser(ctx context.Context, req *protobuf.UserIDRequest) (*protobuf.OrderListResponse, error) {
	logger.Logger.Infof("Received GetOrdersByUser request: UserID=%d", req.UserId)

	var orders []models.Order
	if err := store.DB.Where("user_id = ?", req.UserId).Preload("Items").Find(&orders).Error; err != nil {
		logger.Logger.Errorf("Failed to fetch orders for user from DB: UserID=%d, Error=%v", req.UserId, err)
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
			UserId:      order.UserID,
			Items:       items,
			TotalAmount: float32(order.TotalPrice),
		})
	}

	logger.Logger.Infof("Orders fetched successfully for user: TotalOrders=%d", len(orderResponses))
	return &protobuf.OrderListResponse{Orders: orderResponses}, nil
}
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *protobuf.OrderStatusRequest) (*protobuf.Empty, error) {
	logger.Logger.Infof("Received UpdateOrderStatus request: OrderID=%d, Status=%s", req.OrderId, req.Status)

	var order models.Order
	if err := store.DB.Where("order_id = ?", req.OrderId).First(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to fetch order for status update: OrderID=%d, Error=%v", req.OrderId, err)
		return nil, err
	}

	order.Status = req.Status
	if err := store.DB.Save(&order).Error; err != nil {
		logger.Logger.Errorf("Failed to update order status in DB: OrderID=%d, Error=%v", req.OrderId, err)
		return nil, err
	}

	// // Publish OrderStatusUpdated event to Kafka
	// kafkaReq := &protobuf.PublishRequest{
	// 	Topic:     "Orders",
	// 	EventType: "OrderStatusUpdated",
	// 	Message:   fmt.Sprintf("{\"order_id\": %d, \"status\": \"%s\"}", req.OrderId, req.Status),
	// }

	// _, err := s.KafkaClient.PublishMessage(ctx, kafkaReq)
	// if err != nil {
	// 	logger.Logger.Errorf("Failed to publish OrderStatusUpdated event: %v", err)
	// 	return nil, err
	// }
	// logger.Logger.Infof("OrderStatusUpdated event published successfully: %+v", kafkaReq)

	return &protobuf.Empty{}, nil
}
