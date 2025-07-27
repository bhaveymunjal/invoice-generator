package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config Configuration
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	JWTSecret   string
	ServerPort  string
	ServerMode  string
	CORSOrigins []string
}

func loadConfig() *Config {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "invoice_user"),
		DBPassword:  getEnv("DB_PASSWORD", "invoice_password"),
		DBName:      getEnv("DB_NAME", "invoice_db"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerMode:  getEnv("SERVER_MODE", "debug"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// User Models (same as before)
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Password    string    `json:"-" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	CompanyName string    `json:"company_name"`
	GSTIN       string    `json:"gstin"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Pincode     string    `json:"pincode"`
	Phone       string    `json:"phone"`
	IsAdmin     bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Category struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null;unique"`
	Description string `json:"description"`
	GSTRate     int    `json:"gst_rate" gorm:"not null"`
}

type Item struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `json:"category" gorm:"foreignKey:CategoryID"`
	HSNCode     string    `json:"hsn_code"`
	Unit        string    `json:"unit" gorm:"default:'pcs'"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
	ID             uint              `json:"id" gorm:"primaryKey"`
	InvoiceNumber  string            `json:"invoice_number" gorm:"unique;not null"`
	GeneratedByID  uint              `json:"generated_by_id"`
	GeneratedBy    User              `json:"generated_by" gorm:"foreignKey:GeneratedByID"`
	GeneratedForID uint              `json:"generated_for_id"`
	GeneratedFor   User              `json:"generated_for" gorm:"foreignKey:GeneratedForID"`
	InvoiceType    string            `json:"invoice_type" gorm:"not null;check:invoice_type IN ('CASH','CREDIT','DEBIT')"`
	PaymentStatus  string            `json:"payment_status" gorm:"default:'PENDING';check:payment_status IN ('PENDING','PARTIAL','PAID')"`
	InvoiceDate    time.Time         `json:"invoice_date"`
	DueDate        time.Time         `json:"due_date"`
	SubTotal       float64           `json:"sub_total" gorm:"type:decimal(15,2)"`
	TotalGST       float64           `json:"total_gst" gorm:"type:decimal(15,2)"`
	TotalAmount    float64           `json:"total_amount" gorm:"type:decimal(15,2)"`
	AmountPaid     float64           `json:"amount_paid" gorm:"default:0;type:decimal(15,2)"`
	AmountDue      float64           `json:"amount_due" gorm:"type:decimal(15,2)"`
	Notes          string            `json:"notes"`
	Terms          string            `json:"terms"`
	LineItems      []InvoiceLineItem `json:"line_items" gorm:"foreignKey:InvoiceID"`
	Payments       []Payment         `json:"payments" gorm:"foreignKey:InvoiceID"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type InvoiceLineItem struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	InvoiceID   uint    `json:"invoice_id"`
	ItemID      *uint   `json:"item_id"` // Optional, can be null for custom items
	Item        *Item   `json:"item,omitempty" gorm:"foreignKey:ItemID"`
	Description string  `json:"description" gorm:"not null"`
	Quantity    float64 `json:"quantity" gorm:"type:decimal(10,3)"`
	Rate        float64 `json:"rate" gorm:"type:decimal(15,2)"`
	Amount      float64 `json:"amount" gorm:"type:decimal(15,2)"`
	GSTRate     int     `json:"gst_rate"`
	GSTAmount   float64 `json:"gst_amount" gorm:"type:decimal(15,2)"`
	TotalAmount float64 `json:"total_amount" gorm:"type:decimal(15,2)"`
}

type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	InvoiceID     uint      `json:"invoice_id"`
	Amount        float64   `json:"amount" gorm:"type:decimal(15,2)"`
	PaymentMethod string    `json:"payment_method" gorm:"check:payment_method IN ('CASH','BANK_TRANSFER','CHEQUE','UPI','CARD')"`
	PaymentDate   time.Time `json:"payment_date"`
	Reference     string    `json:"reference"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

// Global variables
var (
	db        *gorm.DB
	config    *Config
	jwtSecret []byte
)

func initDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode)

	var err error
	dbConfig := &gorm.Config{}

	if config.ServerMode == "debug" {
		dbConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err = gorm.Open(postgres.Open(dsn), dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	err = db.AutoMigrate(&User{}, &Category{}, &Item{}, &Invoice{}, &InvoiceLineItem{}, &Payment{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed default data
	seedDefaultCategories()
	log.Println("Database initialized successfully")
}

func seedDefaultCategories() {
	categories := []Category{
		{Name: "Groceries", Description: "Essential food items", GSTRate: 5},
		{Name: "Electronics", Description: "Electronic goods", GSTRate: 18},
		{Name: "Textiles", Description: "Clothing and fabrics", GSTRate: 12},
		{Name: "Services", Description: "Professional services", GSTRate: 18},
		{Name: "Medicines", Description: "Pharmaceutical products", GSTRate: 12},
		{Name: "Books", Description: "Educational materials", GSTRate: 0},
		{Name: "Automobiles", Description: "Vehicles and parts", GSTRate: 28},
		{Name: "Food Items", Description: "Prepared food", GSTRate: 5},
		{Name: "Construction", Description: "Building materials", GSTRate: 18},
		{Name: "Agriculture", Description: "Agricultural products", GSTRate: 0},
	}

	for _, category := range categories {
		var existingCategory Category
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err := db.Create(&category).Error; err != nil {
				log.Printf("Failed to create category %s: %v", category.Name, err)
			}
		}
	}
}

// JWT Claims
type Claims struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(*Claims)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Validation middleware for invoice creation
func validateInvoiceData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice Invoice
		if err := c.ShouldBindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice data: " + err.Error()})
			c.Abort()
			return
		}

		// Validate invoice type
		validTypes := []string{"CASH", "CREDIT", "DEBIT"}
		isValidType := false
		for _, t := range validTypes {
			if invoice.InvoiceType == t {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice type"})
			c.Abort()
			return
		}

		// Validate line items
		if len(invoice.LineItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice must have at least one line item"})
			c.Abort()
			return
		}

		c.Set("invoice", invoice)
		c.Next()
	}
}

// Handlers
func register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email format
	if !strings.Contains(user.Email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Hash password
	if len(user.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	if err := db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
		}
		return
	}

	user.Password = "" // Don't return password
	c.JSON(http.StatusCreated, gin.H{"user": user, "message": "User created successfully"})
}

func login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	claims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

func getUsers(c *gin.Context) {
	var users []User
	query := db.Select("id, email, name, company_name, gstin, address, city, state, pincode, phone, is_admin, created_at, updated_at")

	// If not admin, only return basic info
	isAdmin, _ := c.Get("is_admin")
	if !isAdmin.(bool) {
		query = query.Select("id, name, company_name, email")
	}

	query.Find(&users)
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func getProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user User

	if err := db.Select("id, email, name, company_name, gstin, address, city, state, pincode, phone, is_admin, created_at, updated_at").
		First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func updateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var updateData User

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't allow updating certain fields
	updateData.ID = 0
	updateData.Email = ""
	updateData.Password = ""
	updateData.IsAdmin = false

	if err := db.Model(&User{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func getCategories(c *gin.Context) {
	var categories []Category
	db.Order("name").Find(&categories)
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func createCategory(c *gin.Context) {
	var category Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category name already exists"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create category"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"category": category})
}

func getItems(c *gin.Context) {
	var items []Item
	db.Preload("Category").Order("name").Find(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func createItem(c *gin.Context) {
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate category exists
	var category Category
	if err := db.First(&category, item.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create item"})
		return
	}

	db.Preload("Category").First(&item, item.ID)
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func createInvoice(c *gin.Context) {
	invoice, _ := c.Get("invoice")
	invoiceData := invoice.(Invoice)

	userID, _ := c.Get("user_id")
	invoiceData.GeneratedByID = userID.(uint)

	// Generate invoice number
	invoiceData.InvoiceNumber = generateInvoiceNumber()

	// Set default dates if not provided
	if invoiceData.InvoiceDate.IsZero() {
		invoiceData.InvoiceDate = time.Now()
	}
	if invoiceData.DueDate.IsZero() {
		invoiceData.DueDate = invoiceData.InvoiceDate.AddDate(0, 0, 30) // 30 days default
	}

	// Calculate totals
	calculateInvoiceTotals(&invoiceData)

	// Set amount due
	invoiceData.AmountDue = invoiceData.TotalAmount

	if err := db.Create(&invoiceData).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create invoice: " + err.Error()})
		return
	}

	// Load relationships
	db.Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments").
		First(&invoiceData, invoiceData.ID)

	c.JSON(http.StatusCreated, gin.H{"invoice": invoiceData})
}

func generateInvoiceNumber() string {
	var count int64
	db.Model(&Invoice{}).Count(&count)
	year := time.Now().Year()
	return fmt.Sprintf("INV-%d-%06d", year, count+1)
}

func calculateInvoiceTotals(invoice *Invoice) {
	var subTotal, totalGST float64

	for i := range invoice.LineItems {
		lineItem := &invoice.LineItems[i]
		lineItem.Amount = lineItem.Quantity * lineItem.Rate
		lineItem.GSTAmount = (lineItem.Amount * float64(lineItem.GSTRate)) / 100
		lineItem.TotalAmount = lineItem.Amount + lineItem.GSTAmount

		subTotal += lineItem.Amount
		totalGST += lineItem.GSTAmount
	}

	invoice.SubTotal = subTotal
	invoice.TotalGST = totalGST
	invoice.TotalAmount = subTotal + totalGST
}

func getInvoices(c *gin.Context) {
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	var invoices []Invoice
	query := db.Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments")

	if !isAdmin.(bool) {
		// Regular users can only see invoices they generated or received
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	// Add pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&Invoice{}).Count(&total)

	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices)

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func getInvoice(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	var invoice Invoice
	query := db.Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments")

	if !isAdmin.(bool) {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	if err := query.First(&invoice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice": invoice})
}

func addPayment(c *gin.Context) {
	invoiceID := c.Param("id")
	var payment Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, _ := strconv.Atoi(invoiceID)
	payment.InvoiceID = uint(id)

	// Validate invoice exists and user has access
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	var invoice Invoice
	query := db.Where("id = ?", invoiceID)
	if !isAdmin.(bool) {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	if err := query.First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Validate payment amount
	if payment.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment amount must be positive"})
		return
	}

	if payment.Amount > invoice.AmountDue {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment amount cannot exceed amount due"})
		return
	}

	if err := db.Create(&payment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to add payment"})
		return
	}

	// Update invoice payment status
	updateInvoicePaymentStatus(uint(id))

	c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

func updateInvoicePaymentStatus(invoiceID uint) {
	var invoice Invoice
	db.First(&invoice, invoiceID)

	var totalPaid float64
	db.Model(&Payment{}).Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalPaid)

	invoice.AmountPaid = totalPaid
	invoice.AmountDue = invoice.TotalAmount - totalPaid

	if totalPaid >= invoice.TotalAmount {
		invoice.PaymentStatus = "PAID"
	} else if totalPaid > 0 {
		invoice.PaymentStatus = "PARTIAL"
	} else {
		invoice.PaymentStatus = "PENDING"
	}

	db.Save(&invoice)
}

func getDashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var stats struct {
		TodaySales       float64 `json:"today_sales"`
		TodayCredit      float64 `json:"today_credit"`
		TodayDebit       float64 `json:"today_debit"`
		TotalReceivables float64 `json:"total_receivables"`
		TotalPayables    float64 `json:"total_payables"`
		PendingInvoices  int64   `json:"pending_invoices"`
		TotalInvoices    int64   `json:"total_invoices"`
		ThisMonthSales   float64 `json:"this_month_sales"`
		LastMonthSales   float64 `json:"last_month_sales"`
	}

	baseQuery := db.Model(&Invoice{})
	if !isAdmin.(bool) {
		baseQuery = baseQuery.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	// Today's sales (cash + credit sales where user is generator)
	db.Model(&Invoice{}).Where("generated_by_id = ? AND invoice_date >= ? AND invoice_date < ?",
		userID, today, tomorrow).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TodaySales)

	// Today's credit (invoices generated for others)
	db.Model(&Invoice{}).Where("generated_by_id = ? AND generated_for_id != ? AND invoice_date >= ? AND invoice_date < ?",
		userID, userID, today, tomorrow).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TodayCredit)

	// Today's debit (invoices received from others)
	db.Model(&Invoice{}).Where("generated_for_id = ? AND generated_by_id != ? AND invoice_date >= ? AND invoice_date < ?",
		userID, userID, today, tomorrow).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TodayDebit)

	// Total receivables (what others owe to user)
	db.Model(&Invoice{}).Where("generated_by_id = ? AND generated_for_id != ? AND amount_due > 0",
		userID, userID).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TotalReceivables)

	// Total payables (what user owes to others)
	db.Model(&Invoice{}).Where("generated_for_id = ? AND generated_by_id != ? AND amount_due > 0",
		userID, userID).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TotalPayables)

	// Pending invoices count
	query := db.Model(&Invoice{}).Where("payment_status != 'PAID'")
	if !isAdmin.(bool) {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}
	query.Count(&stats.PendingInvoices)

	// Total invoices count
	query = db.Model(&Invoice{})
	if !isAdmin.(bool) {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}
	query.Count(&stats.TotalInvoices)

	// This month's sales
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	db.Model(&Invoice{}).Where("generated_by_id = ? AND invoice_date >= ?",
		userID, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.ThisMonthSales)

	// Last month's sales
	lastMonth := startOfMonth.AddDate(0, -1, 0)
	db.Model(&Invoice{}).Where("generated_by_id = ? AND invoice_date >= ? AND invoice_date < ?",
		userID, lastMonth, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.LastMonthSales)

	c.JSON(http.StatusOK, gin.H{"dashboard": stats})
}

// Admin specific handlers
func getAdminStats(c *gin.Context) {
	var stats struct {
		TotalUsers    int64   `json:"total_users"`
		TotalInvoices int64   `json:"total_invoices"`
		TotalAmount   float64 `json:"total_amount"`
		PendingAmount float64 `json:"pending_amount"`
		TodayInvoices int64   `json:"today_invoices"`
		TodayAmount   float64 `json:"today_amount"`
	}

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	// Count users
	db.Model(&User{}).Count(&stats.TotalUsers)

	// Count invoices
	db.Model(&Invoice{}).Count(&stats.TotalInvoices)

	// Total amount
	db.Model(&Invoice{}).Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TotalAmount)

	// Pending amount
	db.Model(&Invoice{}).Where("payment_status != 'PAID'").
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.PendingAmount)

	// Today's invoices
	db.Model(&Invoice{}).Where("invoice_date >= ? AND invoice_date < ?", today, tomorrow).
		Count(&stats.TodayInvoices)

	// Today's amount
	db.Model(&Invoice{}).Where("invoice_date >= ? AND invoice_date < ?", today, tomorrow).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TodayAmount)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func deleteInvoice(c *gin.Context) {
	id := c.Param("id")

	// Check if invoice exists and has no payments
	var invoice Invoice
	if err := db.Preload("Payments").First(&invoice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if len(invoice.Payments) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete invoice with payments"})
		return
	}

	// Delete line items first
	db.Where("invoice_id = ?", id).Delete(&InvoiceLineItem{})

	// Delete invoice
	if err := db.Delete(&invoice).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

func main() {
	// Load configuration
	config = loadConfig()
	jwtSecret = []byte(config.JWTSecret)

	// Initialize database
	initDB()

	// Set Gin mode
	gin.SetMode(config.ServerMode)

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request logging middleware
	if config.ServerMode == "debug" {
		r.Use(gin.Logger())
	}

	// Recovery middleware
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", healthCheck)

	// Public routes
	r.POST("/api/register", register)
	r.POST("/api/login", login)

	// Protected routes
	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		// User routes
		api.GET("/profile", getProfile)
		api.PUT("/profile", updateProfile)
		api.GET("/users", getUsers)

		// Category and item routes
		api.GET("/categories", getCategories)
		api.GET("/items", getItems)

		// Invoice routes
		api.GET("/invoices", getInvoices)
		api.GET("/invoices/:id", getInvoice)
		api.POST("/invoices", validateInvoiceData(), createInvoice)
		api.POST("/invoices/:id/payments", addPayment)

		// Dashboard
		api.GET("/dashboard", getDashboard)

		// Admin only routes
		admin := api.Group("/admin")
		admin.Use(adminMiddleware())
		{
			admin.GET("/stats", getAdminStats)
			admin.POST("/categories", createCategory)
			admin.POST("/items", createItem)
			admin.DELETE("/invoices/:id", deleteInvoice)
		}
	}

	// Start server
	port := ":" + config.ServerPort
	log.Printf("Server starting on port %s", port)
	log.Printf("Environment: %s", config.ServerMode)
	log.Printf("CORS Origins: %v", config.CORSOrigins)

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
