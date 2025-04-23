package store

import (
	"e-commerce/logger"
	"e-commerce/models"
	"e-commerce/write-db-service/metrics"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "user=user password=password dbname=ecom host=postgres port=5432 sslmode=disable"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto migrate
	// DB.Migrator().DropTable(&models.OrderItem{}, &models.Order{}, &models.Product{}, &models.Inventory{}, &models.User{}, &models.PaymentDetails{}, &models.Payment{})
	// DB.Migrator().DropTable(&models.Order{})
	err1 := DB.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.Inventory{}, &models.Payment{}, models.PaymentDetails{})

	if err1 != nil {
		logger.Logger.Fatalf("Failed to auto-migrate database schema: %v", err1)
	}

	logger.Logger.Infoln("Connected to the database")
	fmt.Println("store.DB in main:", DB)

}

// QueryDatabase wraps a database query with timing instrumentation
func QueryDatabase(queryName string, queryFunc func() error) error {
	start := time.Now()

	// Execute the query
	err := queryFunc()

	// Record the query duration
	duration := time.Since(start).Seconds()
	metrics.DBQueryDuration.WithLabelValues(queryName).Observe(duration)

	if err != nil {
		logger.Logger.Errorf("Database query failed: %s, Error: %v", queryName, err)
		return err
	}

	logger.Logger.Infof("Database query succeeded: %s, Duration: %.2f seconds", queryName, duration)
	return nil
}
