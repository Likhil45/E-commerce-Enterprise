package services

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/store"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventoryService struct {
	protobuf.UnimplementedInventoryServiceServer
}

// GetStock fetches stock information for a product
func (s *InventoryService) GetStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
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
func (s *InventoryService) UpdateStock(ctx context.Context, req *protobuf.StockUpdateRequest) (*protobuf.StockResponse, error) {
	var inventory models.Inventory

	// Start transaction
	tx := store.DB.Begin()

	// Lock the row to prevent race conditions
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&inventory, "product_id = ?", req.ProductId).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: 0, Status: "Not Found"}, nil
		}
		log.Printf("Error fetching stock for product %d: %v", req.ProductId, err)
		return nil, err
	}

	// Ensure stock is available before updating
	if uint32(inventory.Stock) < req.Quantity {
		tx.Rollback()
		return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: uint32(inventory.Stock), Status: "Insufficient Stock"}, nil
	}

	// Deduct stock
	inventory.Stock -= uint(req.Quantity)
	if err := tx.Save(&inventory).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating stock for product %d: %v", req.ProductId, err)
		return nil, err
	}

	// Commit transaction
	tx.Commit()

	return &protobuf.StockResponse{ProductId: req.ProductId, Quantity: uint32(inventory.Stock), Status: "Updated"}, nil
}

func (s *InventoryService) TrackStock(ctx context.Context, req *protobuf.StockRequest) (*protobuf.StockResponse, error) {
	return nil, nil
}
