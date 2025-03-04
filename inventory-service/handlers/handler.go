package invhandler

import (
	"context"
	"log"

	"e-commerce/protobuf/protobuf"
)

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
	return s.dbClient.GetStock(ctx, req)
}

// UpdateStock modifies inventory on OrderCreated

func (s *InventoryService) UpdateStock(ctx context.Context, req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	// Fetch current stock level
	stock, err := s.dbClient.UpdateStock(ctx, &protobuf.StockUpdateRequest{ProductId: req.ProductId, Quantity: req.Quantity})
	if err != nil {
		log.Println("Error fetching stock:", err)
		return nil, err
	}

	// Check if sufficient stock is available
	if stock.Status == "Insufficient Stock" {
		// // If out of stock, publish OutOfStock event
		// event := map[string]interface{}{
		// 	"product_id": req.ProductId,
		// }
		// eventJSON, err := json.Marshal(event)
		// if err != nil {
		// 	log.Println("Error marshalling OutOfStock event:", err)
		// } else {
		// 	s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{Topic: "Orders", EventType: "OutOfStock", Message: string(eventJSON)})
		// }

		return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: 0, Status: "Out of Stock"}, nil
	}
	// // Send InventoryReserved event
	// event := map[string]interface{}{
	// 	"product_id": req.ProductId,
	// 	"quantity":   req.Quantity,
	// }
	// eventJSON, err := json.Marshal(event)
	// if err != nil {
	// 	log.Println("Error marshalling InventoryReserved event:", err)
	// } else {
	// 	s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "InventoryReserved", Message: string(eventJSON)})
	// }

	// Return updated stock info
	return &protobuf.StockResponse{
		ProductId: req.ProductId,
		Quantity:  stock.Quantity - req.Quantity,
		Status:    "Reserved",
	}, nil
}

func (s *InventoryService) GetStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
	var prod protobuf.StockResponse
	return &prod, nil
}
