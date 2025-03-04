package models

import (
	"time"
)

type Order struct {
	OrderID    uint32      `gorm:"primaryKey;autoIncrement"`
	UserID     uint32      `gorm:"not null"`
	OrderDate  time.Time   `gorm:"default:CURRENT_TIMESTAMP"`
	Status     string      `gorm:"size:50;default:'pending'"` // "pending", "confirmed", "shipped", "cancelled"
	TotalPrice float64     `gorm:"type:decimal(10,2)"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	OrderItemsID uint32  `gorm:"primaryKey;autoIncrement"`
	OrderID      uint32  `gorm:"not null;index"`
	ProductID    uint32  `gorm:"not null"`
	Quantity     uint    `gorm:"not null"`
	Price        float64 `gorm:"type:decimal(10,2)"`
}
