package models

import (
	"time"
)

type Payment struct {
	ID        uint    `gorm:"primaryKey"`
	OrderID   uint    `gorm:"not null;uniqueIndex"`
	Amount    float64 `gorm:"type:decimal(10,2)"`
	Status    string  `gorm:"size:50;default:'pending'"` // "pending", "completed", "failed"
	CreatedAt time.Time
}
