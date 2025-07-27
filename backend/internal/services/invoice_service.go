package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"invoice-generator/internal/constants"
	"invoice-generator/internal/database"
	"invoice-generator/internal/models"
)

// InvoiceService handles invoice-related business logic
type InvoiceService struct{}

// NewInvoiceService creates a new invoice service
func NewInvoiceService() *InvoiceService {
	return &InvoiceService{}
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(invoice *models.Invoice, userID uint) error {
	invoice.GeneratedByID = userID

	// Generate invoice number
	invoice.InvoiceNumber = s.generateInvoiceNumber()

	// Set default dates if not provided
	if invoice.InvoiceDate.IsZero() {
		invoice.InvoiceDate = time.Now()
	}
	if invoice.DueDate.IsZero() {
		invoice.DueDate = invoice.InvoiceDate.AddDate(0, 0, constants.DefaultDueDays)
	}

	// Calculate totals
	s.calculateInvoiceTotals(invoice)

	// Set amount due
	invoice.AmountDue = invoice.TotalAmount

	if err := database.GetDB().Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Load relationships
	database.GetDB().Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments").
		First(invoice, invoice.ID)

	return nil
}

// GetInvoices returns invoices with pagination and access control
func (s *InvoiceService) GetInvoices(userID uint, isAdmin bool, page, limit int) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	query := database.GetDB().Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments")

	if !isAdmin {
		// Regular users can only see invoices they generated or received
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	// Count total
	var total int64
	query.Model(&models.Invoice{}).Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetInvoice returns a single invoice by ID with access control
func (s *InvoiceService) GetInvoice(id uint, userID uint, isAdmin bool) (*models.Invoice, error) {
	var invoice models.Invoice
	query := database.GetDB().Preload("GeneratedBy").Preload("GeneratedFor").
		Preload("LineItems.Item.Category").Preload("Payments")

	if !isAdmin {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	if err := query.First(&invoice, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	return &invoice, nil
}

// AddPayment adds a payment to an invoice
func (s *InvoiceService) AddPayment(invoiceID uint, payment *models.Payment, userID uint, isAdmin bool) error {
	payment.InvoiceID = invoiceID

	// Validate invoice exists and user has access
	var invoice models.Invoice
	query := database.GetDB().Where("id = ?", invoiceID)
	if !isAdmin {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}

	if err := query.First(&invoice).Error; err != nil {
		return errors.New("invoice not found")
	}

	// Validate payment amount
	if payment.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}

	if payment.Amount > invoice.AmountDue {
		return errors.New("payment amount cannot exceed amount due")
	}

	if err := database.GetDB().Create(payment).Error; err != nil {
		return errors.New("failed to add payment")
	}

	// Update invoice payment status
	s.updateInvoicePaymentStatus(invoiceID)

	return nil
}

// DeleteInvoice deletes an invoice (admin only)
func (s *InvoiceService) DeleteInvoice(id uint) error {
	// Check if invoice exists and has no payments
	var invoice models.Invoice
	if err := database.GetDB().Preload("Payments").First(&invoice, id).Error; err != nil {
		return errors.New("invoice not found")
	}

	if len(invoice.Payments) > 0 {
		return errors.New("cannot delete invoice with payments")
	}

	// Delete line items first
	database.GetDB().Where("invoice_id = ?", id).Delete(&models.InvoiceLineItem{})

	// Delete invoice
	if err := database.GetDB().Delete(&invoice).Error; err != nil {
		return errors.New("failed to delete invoice")
	}

	return nil
}

// generateInvoiceNumber generates a unique invoice number
func (s *InvoiceService) generateInvoiceNumber() string {
	var count int64
	database.GetDB().Model(&models.Invoice{}).Count(&count)
	year := time.Now().Year()
	return fmt.Sprintf("INV-%d-%06d", year, count+1)
}

// calculateInvoiceTotals calculates invoice totals
func (s *InvoiceService) calculateInvoiceTotals(invoice *models.Invoice) {
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

// updateInvoicePaymentStatus updates the payment status of an invoice
func (s *InvoiceService) updateInvoicePaymentStatus(invoiceID uint) {
	var invoice models.Invoice
	database.GetDB().First(&invoice, invoiceID)

	var totalPaid float64
	database.GetDB().Model(&models.Payment{}).Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalPaid)

	invoice.AmountPaid = totalPaid
	invoice.AmountDue = invoice.TotalAmount - totalPaid

	if totalPaid >= invoice.TotalAmount {
		invoice.PaymentStatus = constants.PaymentStatusPaid
	} else if totalPaid > 0 {
		invoice.PaymentStatus = constants.PaymentStatusPartial
	} else {
		invoice.PaymentStatus = constants.PaymentStatusPending
	}

	database.GetDB().Save(&invoice)
}
