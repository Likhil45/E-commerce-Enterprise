package services

import (
	"context"
	"log"

	"e-commerce/models"
	"e-commerce/protobuf/protobuf"

	"e-commerce/write-db-service/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DatabaseService struct {
	protobuf.UnimplementedDatabaseServiceServer
	kafkaClient protobuf.KafkaProducerServiceClient
}

func NewDatabaseServiceClient(kafkaClient protobuf.KafkaProducerServiceClient) *DatabaseService {
	return &DatabaseService{kafkaClient: kafkaClient}
}

// GetStock fetches stock information for a product
func (s *DatabaseService) GetStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
	var inventory models.Inventory
	result := store.DB.First(&inventory, "product_id = ?", req.ProductId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: 0, Status: "Not Found"}, nil
		}
		log.Printf("Error fetching stock for product %d: %v", req.ProductId, result.Error)
		return nil, result.Error
	}

	return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: uint32(inventory.Stock), Status: "Available"}, nil
}

// UpdateStock updates the stock quantity when an order is placed
func (s *DatabaseService) UpdateStock(ctx context.Context, req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	var product models.Product

	// Start transaction
	tx := store.DB.Begin()

	// Lock the row to prevent race conditions
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, "product_id = ?", req.ProductId).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: 0, Status: "Not Found"}, nil
		}
		log.Printf("Error fetching stock for product %d: %v", req.ProductId, err)
		return nil, err
	}

	// Ensure stock is available before updating
	if uint32(product.StockQuantity) < req.Quantity {
		tx.Rollback()

		// event := map[string]interface{}{
		// 	"product_id": req.ProductId,
		// }

		// eventJSON, err := json.Marshal(event)
		// if err != nil {
		// 	log.Println("Error marshalling OutOfStock event:", err)
		// }

		// s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{Topic: "Orders", EventType: "OutOfStock", Message: string(eventJSON)})

		return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: uint32(product.StockQuantity), Status: "Insufficient Stock"}, nil

	}

	// Deduct stock
	product.StockQuantity -= uint(req.Quantity)
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating stock for product %d: %v", req.ProductId, err)
		return nil, err
	}

	// Commit transaction
	tx.Commit()

	// Send InventoryReserved event

	// event := map[string]interface{}{
	// 	"product_id": req.ProductId,
	// 	"quantity":   req.Quantity,
	// }

	// eventJSON, err := json.Marshal(&event)
	// if err != nil {
	// 	log.Println("Error marshalling InventoryReserved event:", err)
	// } else {
	// 	s.kafkaClient.PublishMessage(ctx, &protobuf.PublishRequest{EventType: "InventoryReserved", Message: string(eventJSON)})
	// }
	return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: req.Quantity, Status: "InventoryReserved"}, nil

}
