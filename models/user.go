package models

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"not null"`        // e.g., "admin", "customer"
	Provider  string `gorm:"default:'local'"` // "local", "google", "facebook", etc.
	CreatedAt time.Time
}
