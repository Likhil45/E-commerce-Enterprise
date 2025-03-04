package models

import "time"

type Product struct {
	ProductID     uint    `gorm:"primaryKey"`
	Name          string  `gorm:"size:255;not null"`
	Description   string  `gorm:"size:500"`
	Price         float64 `gorm:"type:decimal(10,2);not null"`
	StockQuantity uint    `gorm:"not null"`
	Category      string  `gorm:"size:100"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
