package models

import (
	"time"
)

// User represents a user in the system
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

// Category represents a product/service category
type Category struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null;unique"`
	Description string `json:"description"`
	GSTRate     int    `json:"gst_rate" gorm:"not null"`
}

// Item represents a product or service
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

// Invoice represents an invoice
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

// InvoiceLineItem represents a line item in an invoice
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

// Payment represents a payment made against an invoice
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

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
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

// AdminStats represents admin dashboard statistics
type AdminStats struct {
	TotalUsers    int64   `json:"total_users"`
	TotalInvoices int64   `json:"total_invoices"`
	TotalAmount   float64 `json:"total_amount"`
	PendingAmount float64 `json:"pending_amount"`
	TodayInvoices int64   `json:"today_invoices"`
	TodayAmount   float64 `json:"today_amount"`
}
