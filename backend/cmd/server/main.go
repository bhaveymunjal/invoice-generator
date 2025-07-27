package main

import (
	"log"

	"invoice-generator/internal/config"
	"invoice-generator/internal/database"
	"invoice-generator/internal/handlers"
	"invoice-generator/internal/routes"
	"invoice-generator/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	jwtSecret := []byte(cfg.JWTSecret)

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize services
	userService := services.NewUserService(jwtSecret)
	invoiceService := services.NewInvoiceService()
	dashboardService := services.NewDashboardService()
	catalogService := services.NewCatalogService()

	// Initialize handlers
	h := handlers.NewHandlers(
		userService,
		invoiceService,
		dashboardService,
		catalogService,
	)

	// Setup routes
	r := routes.SetupRoutes(cfg, h, jwtSecret)

	// Start server
	port := ":" + cfg.ServerPort
	log.Printf("Server starting on port %s", port)
	log.Printf("Environment: %s", cfg.ServerMode)
	log.Printf("CORS Origins: %v", cfg.CORSOrigins)

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
