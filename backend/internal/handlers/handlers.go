package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"invoice-generator/internal/constants"
	"invoice-generator/internal/models"
	"invoice-generator/internal/services"
)

// Handlers holds all the service dependencies
type Handlers struct {
	userService      *services.UserService
	invoiceService   *services.InvoiceService
	dashboardService *services.DashboardService
	catalogService   *services.CatalogService
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	userService *services.UserService,
	invoiceService *services.InvoiceService,
	dashboardService *services.DashboardService,
	catalogService *services.CatalogService,
) *Handlers {
	return &Handlers{
		userService:      userService,
		invoiceService:   invoiceService,
		dashboardService: dashboardService,
		catalogService:   catalogService,
	}
}

// Auth Handlers

// Register handles user registration
func (h *Handlers) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Register(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user, "message": "User created successfully"})
}

// Login handles user login
func (h *Handlers) Login(c *gin.Context) {
	var loginData models.LoginRequest
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.userService.Login(&loginData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// User Handlers

// GetUsers returns all users
func (h *Handlers) GetUsers(c *gin.Context) {
	isAdmin, _ := c.Get("is_admin")
	users, err := h.userService.GetUsers(isAdmin.(bool))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetProfile returns user profile
func (h *Handlers) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.userService.GetProfile(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateProfile updates user profile
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var updateData models.User

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateProfile(userID.(uint), &updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// Catalog Handlers

// GetCategories returns all categories
func (h *Handlers) GetCategories(c *gin.Context) {
	categories, err := h.catalogService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// CreateCategory creates a new category
func (h *Handlers) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.catalogService.CreateCategory(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"category": category})
}

// GetItems returns all items
func (h *Handlers) GetItems(c *gin.Context) {
	items, err := h.catalogService.GetItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateItem creates a new item
func (h *Handlers) CreateItem(c *gin.Context) {
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.catalogService.CreateItem(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item})
}

// Invoice Handlers

// CreateInvoice creates a new invoice
func (h *Handlers) CreateInvoice(c *gin.Context) {
	invoice, _ := c.Get("invoice")
	invoiceData := invoice.(models.Invoice)
	userID, _ := c.Get("user_id")

	if err := h.invoiceService.CreateInvoice(&invoiceData, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"invoice": invoiceData})
}

// GetInvoices returns invoices with pagination
func (h *Handlers) GetInvoices(c *gin.Context) {
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(constants.DefaultPageLimit)))

	invoices, total, err := h.invoiceService.GetInvoices(userID.(uint), isAdmin.(bool), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetInvoice returns a single invoice
func (h *Handlers) GetInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	invoice, err := h.invoiceService.GetInvoice(uint(id), userID.(uint), isAdmin.(bool))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice": invoice})
}

// AddPayment adds a payment to an invoice
func (h *Handlers) AddPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	if err := h.invoiceService.AddPayment(uint(id), &payment, userID.(uint), isAdmin.(bool)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

// DeleteInvoice deletes an invoice (admin only)
func (h *Handlers) DeleteInvoice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	if err := h.invoiceService.DeleteInvoice(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

// Dashboard Handlers

// GetDashboard returns dashboard statistics
func (h *Handlers) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	stats, err := h.dashboardService.GetDashboardStats(userID.(uint), isAdmin.(bool))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dashboard": stats})
}

// GetAdminStats returns admin dashboard statistics
func (h *Handlers) GetAdminStats(c *gin.Context) {
	stats, err := h.dashboardService.GetAdminStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// HealthCheck returns health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}
