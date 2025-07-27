package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"invoice-generator/internal/config"
	"invoice-generator/internal/constants"
	"invoice-generator/internal/models"
)

// DB holds the database connection
var DB *gorm.DB

// Initialize initializes the database connection and runs migrations
func Initialize(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	var err error
	dbConfig := &gorm.Config{}

	if cfg.ServerMode == "debug" {
		dbConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	DB, err = gorm.Open(postgres.Open(dsn), dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Item{},
		&models.Invoice{},
		&models.InvoiceLineItem{},
		&models.Payment{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed default data
	if err := seedDefaultCategories(); err != nil {
		return fmt.Errorf("failed to seed default categories: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// seedDefaultCategories seeds the database with default categories
func seedDefaultCategories() error {
	categories := []models.Category{
		{Name: "Groceries", Description: "Essential food items", GSTRate: constants.DefaultGSTRateFive},
		{Name: "Electronics", Description: "Electronic goods", GSTRate: constants.DefaultGSTRateEighteen},
		{Name: "Textiles", Description: "Clothing and fabrics", GSTRate: constants.DefaultGSTRateTwelve},
		{Name: "Services", Description: "Professional services", GSTRate: constants.DefaultGSTRateEighteen},
		{Name: "Medicines", Description: "Pharmaceutical products", GSTRate: constants.DefaultGSTRateTwelve},
		{Name: "Books", Description: "Educational materials", GSTRate: constants.DefaultGSTRateZero},
		{Name: "Automobiles", Description: "Vehicles and parts", GSTRate: constants.DefaultGSTRateTwentyEight},
		{Name: "Food Items", Description: "Prepared food", GSTRate: constants.DefaultGSTRateFive},
		{Name: "Construction", Description: "Building materials", GSTRate: constants.DefaultGSTRateEighteen},
		{Name: "Agriculture", Description: "Agricultural products", GSTRate: constants.DefaultGSTRateZero},
	}

	for _, category := range categories {
		var existingCategory models.Category
		if err := DB.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err := DB.Create(&category).Error; err != nil {
				log.Printf("Failed to create category %s: %v", category.Name, err)
			}
		}
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
