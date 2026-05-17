package services

import (
	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type DashboardStats struct {
	ItemsCount     int64
	InvoicesCount  int64
	TreasuryTotal  float64
	CustomersCount int64
	SuppliersCount int64
	PurchasesCount int64
	SalesTotal     float64
	Notifications  []models.Notification
}

type DashboardService struct{ db *gorm.DB }

func NewDashboardService(db *gorm.DB) *DashboardService { return &DashboardService{db: db} }

func (s *DashboardService) Stats() DashboardStats {
	var stats DashboardStats
	s.db.Model(&models.Item{}).Count(&stats.ItemsCount)
	s.db.Model(&models.SalesInvoice{}).Count(&stats.InvoicesCount)
	s.db.Model(&models.Customer{}).Count(&stats.CustomersCount)
	s.db.Model(&models.Supplier{}).Count(&stats.SuppliersCount)
	s.db.Model(&models.PurchaseInvoice{}).Count(&stats.PurchasesCount)
	s.db.Model(&models.Treasury{}).Select("COALESCE(SUM(balance), 0)").Scan(&stats.TreasuryTotal)
	s.db.Model(&models.SalesInvoice{}).Select("COALESCE(SUM(total), 0)").Scan(&stats.SalesTotal)
	s.db.Order("created_at desc").Limit(10).Find(&stats.Notifications)
	return stats
}
