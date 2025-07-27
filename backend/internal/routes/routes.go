package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"invoice-generator/internal/config"
	"invoice-generator/internal/handlers"
	"invoice-generator/internal/middleware"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(cfg *config.Config, h *handlers.Handlers, jwtSecret []byte) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.ServerMode)

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request logging middleware
	if cfg.ServerMode == "debug" {
		r.Use(gin.Logger())
	}

	// Recovery middleware
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", h.HealthCheck)

	// Public routes
	r.POST("/api/register", h.Register)
	r.POST("/api/login", h.Login)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// User routes
		api.GET("/profile", h.GetProfile)
		api.PUT("/profile", h.UpdateProfile)
		api.GET("/users", h.GetUsers)

		// Category and item routes
		api.GET("/categories", h.GetCategories)
		api.GET("/items", h.GetItems)

		// Invoice routes
		api.GET("/invoices", h.GetInvoices)
		api.GET("/invoices/:id", h.GetInvoice)
		api.POST("/invoices", middleware.ValidateInvoiceData(), h.CreateInvoice)
		api.POST("/invoices/:id/payments", h.AddPayment)

		// Dashboard
		api.GET("/dashboard", h.GetDashboard)

		// Admin only routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/stats", h.GetAdminStats)
			admin.POST("/categories", h.CreateCategory)
			admin.POST("/items", h.CreateItem)
			admin.DELETE("/invoices/:id", h.DeleteInvoice)
		}
	}

	return r
}
