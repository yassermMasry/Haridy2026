package database

import (
	"haridy2026/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Branch{},
		&models.Warehouse{},
		&models.User{},
		&models.Customer{},
		&models.CustomerTransaction{},
		&models.Supplier{},
		&models.SupplierTransaction{},
		&models.ItemCategory{},
		&models.Item{},
		&models.ItemWarehouseBalance{},
		&models.WarehouseTransfer{},
		&models.StockMovement{},
		&models.SalesInvoice{},
		&models.SalesInvoiceItem{},
		&models.Treasury{},
		&models.TreasuryTransaction{},
		&models.PurchaseInvoice{},
		&models.PurchaseInvoiceItem{},
		&models.ChartOfAccount{},
		&models.JournalEntry{},
		&models.JournalEntryLine{},
		&models.AuditLog{},
		&models.RBACRole{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},
		&models.SalesReturn{},
		&models.SalesReturnItem{},
		&models.PurchaseReturn{},
		&models.PurchaseReturnItem{},
		&models.ReceiptVoucher{},
		&models.PaymentVoucher{},
		&models.Notification{},
		&models.LoginAttempt{},
	)
}
