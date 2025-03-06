package models

import (
	"time"
)

type User struct {
	ID        string          `gorm:"primaryKey"`
	Username  string          `gorm:"unique;not null"`
	Email     string          `gorm:"unique;not null"`
	Password  string          `gorm:"not null"`
	Role      string          `gorm:"not null"`              // e.g., "admin", "customer"
	Provider  string          `gorm:"default:'local'"`       // "local", "google", "facebook", etc.
	Balance   float64         `gorm:"not null;default:0.00"` // User wallet balance
	Payment   *PaymentDetails `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
}

type PaymentDetails struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	UserID        string `gorm:"not null"`                // One-to-one relationship with User
	PaymentMethod string `gorm:"not null;default:'card'"` // "card", "paypal", "crypto", etc.
	CardNumber    string `gorm:"size:20"`                 // Masked card number (e.g., "**** **** **** 1234")
	ExpiryDate    string `gorm:"size:7"`                  // Format: "MM/YY"
	CVV           string `gorm:"size:4"`                  // Encrypted CVV (if stored, use secure storage)
}
