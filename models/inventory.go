package models

import "time"

type Inventory struct {
	ID        uint   `gorm:"primaryKey"`
	ProductID uint   `gorm:"uniqueIndex;not null"`
	Stock     uint   `gorm:"not null"`
	Warehouse string `gorm:"size:100"`
	UpdatedAt time.Time
}
