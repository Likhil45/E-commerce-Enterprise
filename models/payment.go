package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Payment struct {
	gorm.Model
	PaymentID     int       `gorm:"primaryKey;unique"`
	OrderID       int       `gorm:"not null;unique"`
	PaymentDate   time.Time `gorm:"not null"`
	Amount        float64   `gorm:"not null"`
	PaymentMethod string    `gorm:"not null"`
	Status        string    `gorm:"not null"`
	Order         Order     `gorm:"foreignKey:OrderID"`
}
