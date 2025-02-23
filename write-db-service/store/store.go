package store

import (
	"e-commerce/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "user=user password=password dbname=ecom port=5432 sslmode=disable"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto migrate
	err = DB.AutoMigrate(&models.User{}, &models.Product{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to the database")
	fmt.Println("store.DB in main:", DB)

}
