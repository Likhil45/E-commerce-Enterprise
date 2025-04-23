package invhandler

import (
	"context"
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
)

var stockUpdatesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "inventory_stock_updates_total",
		Help: "Total number of stock updates processed",
	},
	[]string{"status"},
)

func Init() {
	prometheus.MustRegister(stockUpdatesTotal)
}

type InventoryService struct {
	protobuf.UnimplementedInventoryServiceServer
	dbClient    protobuf.DatabaseServiceClient
	kafkaClient protobuf.KafkaProducerServiceClient
}

func NewInventoryService(dbClient protobuf.DatabaseServiceClient, kafkaClient protobuf.KafkaProducerServiceClient) *InventoryService {
	return &InventoryService{dbClient: dbClient, kafkaClient: kafkaClient}
}

// TrackStock fetches stock levels
func (s *InventoryService) TrackStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
	logger.Logger.Infof("TrackStock called for ProductID=%d", req.ProductId)
	return s.dbClient.GetStock(ctx, req)
}

// UpdateStock modifies inventory on OrderCreated
func (s *InventoryService) UpdateStock(ctx context.Context, req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	logger.Logger.Infof("Updating stock for ProductID=%d, Quantity=%d", req.ProductId, req.Quantity)

	// Fetch current stock level
	stock, err := s.dbClient.UpdateStock(ctx, &protobuf.StockUpdateRequest{ProductId: req.ProductId, Quantity: req.Quantity})
	if err != nil {
		logger.Logger.Errorf("Error fetching stock for ProductID=%d: %v", req.ProductId, err)
		stockUpdatesTotal.WithLabelValues("error").Inc() // Increment error counter
		return nil, err
	}

	// Check if sufficient stock is available
	if stock.Status == "Insufficient Stock" {
		logger.Logger.Warnf("Insufficient stock for ProductID=%d", req.ProductId)
		stockUpdatesTotal.WithLabelValues("insufficient_stock").Inc() // Increment insufficient stock counter

		// If out of stock, publish OutOfStock event
		event := map[string]interface{}{
			"product_id": req.ProductId,
		}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			logger.Logger.Errorf("Error marshalling OutOfStock event for ProductID=%d: %v", req.ProductId, err)
		} else {
			s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{
				Topic:     "Orders",
				EventType: "OutOfStock",
				Message:   string(eventJSON),
			})
			logger.Logger.Infof("OutOfStock event published for ProductID=%d", req.ProductId)
		}

		return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: 0, Status: "Out of Stock"}, nil
	}

	// Send InventoryReserved event
	event := map[string]interface{}{
		"product_id": req.ProductId,
		"quantity":   req.Quantity,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		logger.Logger.Errorf("Error marshalling InventoryReserved event for ProductID=%d: %v", req.ProductId, err)
	} else {
		s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{
			EventType: "InventoryReserved",
			Message:   string(eventJSON),
		})
		logger.Logger.Infof("InventoryReserved event published for ProductID=%d", req.ProductId)
	}

	logger.Logger.Infof("Stock updated successfully for ProductID=%d", req.ProductId)
	stockUpdatesTotal.WithLabelValues("success").Inc()

	// Return updated stock info
	return &protobuf.StockResponse{
		ProductId: req.ProductId,
		Quantity:  stock.Quantity - req.Quantity,
		Status:    "Reserved",
	}, nil
}

func (s *InventoryService) GetStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
	logger.Logger.Infof("GetStock called for ProductID=%d", req.ProductId)
	var prod protobuf.StockResponse
	return &prod, nil
}
