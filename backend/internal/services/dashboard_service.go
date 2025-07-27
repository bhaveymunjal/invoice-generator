package services

import (
	"time"

	"invoice-generator/internal/database"
	"invoice-generator/internal/models"
)

// DashboardService handles dashboard-related business logic
type DashboardService struct{}

// NewDashboardService creates a new dashboard service
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetDashboardStats returns dashboard statistics for a user
func (s *DashboardService) GetDashboardStats(userID uint, isAdmin bool) (*models.DashboardStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	stats := &models.DashboardStats{}

	// Today's sales (cash + credit sales where user is generator)
	database.GetDB().Model(&models.Invoice{}).Where("generated_by_id = ? AND invoice_date >= ? AND invoice_date < ?",
		userID, today, tomorrow).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TodaySales)

	// Today's credit (invoices generated for others)
	database.GetDB().Model(&models.Invoice{}).Where("generated_by_id = ? AND generated_for_id != ? AND invoice_date >= ? AND invoice_date < ?",
		userID, userID, today, tomorrow).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TodayCredit)

	// Today's debit (invoices received from others)
	database.GetDB().Model(&models.Invoice{}).Where("generated_for_id = ? AND generated_by_id != ? AND invoice_date >= ? AND invoice_date < ?",
		userID, userID, today, tomorrow).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TodayDebit)

	// Total receivables (what others owe to user)
	database.GetDB().Model(&models.Invoice{}).Where("generated_by_id = ? AND generated_for_id != ? AND amount_due > 0",
		userID, userID).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TotalReceivables)

	// Total payables (what user owes to others)
	database.GetDB().Model(&models.Invoice{}).Where("generated_for_id = ? AND generated_by_id != ? AND amount_due > 0",
		userID, userID).
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.TotalPayables)

	// Pending invoices count
	query := database.GetDB().Model(&models.Invoice{}).Where("payment_status != 'PAID'")
	if !isAdmin {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}
	query.Count(&stats.PendingInvoices)

	// Total invoices count
	query = database.GetDB().Model(&models.Invoice{})
	if !isAdmin {
		query = query.Where("generated_by_id = ? OR generated_for_id = ?", userID, userID)
	}
	query.Count(&stats.TotalInvoices)

	// This month's sales
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	database.GetDB().Model(&models.Invoice{}).Where("generated_by_id = ? AND invoice_date >= ?",
		userID, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.ThisMonthSales)

	// Last month's sales
	lastMonth := startOfMonth.AddDate(0, -1, 0)
	database.GetDB().Model(&models.Invoice{}).Where("generated_by_id = ? AND invoice_date >= ? AND invoice_date < ?",
		userID, lastMonth, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.LastMonthSales)

	return stats, nil
}

// GetAdminStats returns admin dashboard statistics
func (s *DashboardService) GetAdminStats() (*models.AdminStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	stats := &models.AdminStats{}

	// Count users
	database.GetDB().Model(&models.User{}).Count(&stats.TotalUsers)

	// Count invoices
	database.GetDB().Model(&models.Invoice{}).Count(&stats.TotalInvoices)

	// Total amount
	database.GetDB().Model(&models.Invoice{}).Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TotalAmount)

	// Pending amount
	database.GetDB().Model(&models.Invoice{}).Where("payment_status != 'PAID'").
		Select("COALESCE(SUM(amount_due), 0)").Row().Scan(&stats.PendingAmount)

	// Today's invoices
	database.GetDB().Model(&models.Invoice{}).Where("invoice_date >= ? AND invoice_date < ?", today, tomorrow).
		Count(&stats.TodayInvoices)

	// Today's amount
	database.GetDB().Model(&models.Invoice{}).Where("invoice_date >= ? AND invoice_date < ?", today, tomorrow).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&stats.TodayAmount)

	return stats, nil
}
