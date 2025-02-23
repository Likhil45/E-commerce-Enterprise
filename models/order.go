package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Order struct {
	gorm.Model
	OrderID     int       `gorm:"primaryKey;unique"`
	UserID      int       `gorm:"not null;unique"`
	OrderDate   time.Time `gorm:"not null"`
	Status      string    `gorm:"not null"`
	TotalAmount float64   `gorm:"not null"`
	User        User      `gorm:"foreignKey:UserID"`
}
