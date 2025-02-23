package models

import "time"

type Product struct {
	ProductID   int       `json:"product_id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
}
