package services

import (
	"context"
	"e-commerce/models"
	"e-commerce/protobuf/protobuf"
	"e-commerce/write-db-service/store"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WriteProduct struct {
	protobuf.UnimplementedProductServiceServer
}

func (p *WriteProduct) CreateProduct(ctx context.Context, req *protobuf.ProductRequest) (*protobuf.ProductResponse, error) {

	var prod models.Product
	prod.Name = req.GetName()
	prod.Description = req.GetDescription()
	prod.Price = float64(req.GetPrice())
	prod.ProductID = uint(req.GetProductId())
	prod.StockQuantity = uint(req.Quantity)

	if err := store.DB.Create(&prod).Error; err != nil {
		log.Printf("Error while creating user: %v", err)
		return nil, err
	}
	return &protobuf.ProductResponse{Name: prod.Name, Description: prod.Description, Price: float32(prod.Price), ProductId: uint32(prod.ProductID), Quantity: uint32(prod.StockQuantity)}, nil

}

func (p *WriteProduct) GetProduct(ctx context.Context, req *protobuf.ProductIDRequest) (*protobuf.ProductResponse, error) {
	var prod models.Product
	if err := store.DB.Where("product_id = ?", req.GetProductId()).First(&prod).Error; err != nil {
		return nil, err
	}
	return &protobuf.ProductResponse{ProductId: uint32(prod.ProductID), Price: float32(prod.Price), Name: prod.Name, Description: prod.Description}, nil
}

func (p *WriteProduct) DeleteProduct(ctx context.Context, req *protobuf.ProductIDRequest) (*protobuf.Empty, error) {

	var product models.Product
	if err := store.DB.Where("product_id = ?", req.GetProductId()).First(&product).Error; err != nil {

		return nil, status.Errorf(codes.Internal, "Database error: %v", err)
	}

	// Delete the product
	if err := store.DB.Delete(&product).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete product: %v", err)
	}

	// Return an empty response on success
	return &protobuf.Empty{}, nil
}
func (p *WriteProduct) ListProducts(ctx context.Context, req *protobuf.Empty) (*protobuf.ProductListResponse, error) {
	var prods []models.Product
	if err := store.DB.Find(&prods).Error; err != nil {
		return nil, err
	}
	var productResponses []*protobuf.ProductResponse
	for _, prod := range prods {
		productResponses = append(productResponses, &protobuf.ProductResponse{
			ProductId:   uint32(prod.ProductID),
			Name:        prod.Name,
			Description: prod.Description,
			Price:       float32(prod.Price),
			Quantity:    uint32(prod.StockQuantity),
		})
	}
	return &protobuf.ProductListResponse{Products: productResponses}, nil
}

func (p *WriteProduct) UpdateProduct(ctx context.Context, req *protobuf.ProductRequest) (*protobuf.ProductResponse, error) {
	var product models.Product

	// Check if the product exists
	if err := store.DB.Where("product_id = ?", req.GetProductId()).First(&product).Error; err != nil {
		return nil, err
	}

	// Update the product fields
	product.Name = req.GetName()
	product.Description = req.GetDescription()
	product.Price = float64(req.GetPrice())

	// Save the updated product
	if err := store.DB.Save(&product).Error; err != nil {
		return nil, err
	}

	// Return the updated product response
	return &protobuf.ProductResponse{
		ProductId:   uint32(product.ProductID),
		Name:        product.Name,
		Description: product.Description,
		Price:       float32(product.Price),
	}, nil
}
