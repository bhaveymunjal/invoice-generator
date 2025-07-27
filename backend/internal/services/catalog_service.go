package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"invoice-generator/internal/database"
	"invoice-generator/internal/models"
)

// CatalogService handles categories and items business logic
type CatalogService struct{}

// NewCatalogService creates a new catalog service
func NewCatalogService() *CatalogService {
	return &CatalogService{}
}

// GetCategories returns all categories
func (s *CatalogService) GetCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := database.GetDB().Order("name").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory creates a new category
func (s *CatalogService) CreateCategory(category *models.Category) error {
	if err := database.GetDB().Create(category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("category name already exists")
		}
		return errors.New("failed to create category")
	}
	return nil
}

// GetItems returns all items with their categories
func (s *CatalogService) GetItems() ([]models.Item, error) {
	var items []models.Item
	if err := database.GetDB().Preload("Category").Order("name").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CreateItem creates a new item
func (s *CatalogService) CreateItem(item *models.Item) error {
	// Validate category exists
	var category models.Category
	if err := database.GetDB().First(&category, item.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid category")
		}
		return err
	}

	if err := database.GetDB().Create(item).Error; err != nil {
		return errors.New("failed to create item")
	}

	// Load the category relationship
	database.GetDB().Preload("Category").First(item, item.ID)
	return nil
}
