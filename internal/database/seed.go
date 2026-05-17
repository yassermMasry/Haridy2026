package database

import (
	"haridy2026/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	var branch models.Branch
	if err := db.Where("code = ?", "MAIN").FirstOrCreate(&branch, models.Branch{Name: "Main Branch", Code: "MAIN", IsActive: true}).Error; err != nil {
		return err
	}
	var warehouse models.Warehouse
	if err := db.Where("code = ?", "MAIN-WH").FirstOrCreate(&warehouse, models.Warehouse{Name: "Main Warehouse", Code: "MAIN-WH", BranchID: branch.ID, IsActive: true}).Error; err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var admin models.User
	if err := db.Where("username = ?", "admin").FirstOrCreate(&admin, models.User{
		Username:        "admin",
		PasswordHash:    string(hash),
		Role:            models.RoleAdmin,
		CurrentBranchID: &branch.ID,
	}).Error; err != nil {
		return err
	}

	var treasury models.Treasury
	if err := db.Where("name = ?", "الخزينة الرئيسية").FirstOrCreate(&treasury, models.Treasury{Name: "الخزينة الرئيسية"}).Error; err != nil {
		return err
	}
	var category models.ItemCategory
	if err := db.Where("name = ?", "عام").FirstOrCreate(&category, models.ItemCategory{Name: "عام"}).Error; err != nil {
		return err
	}
	accounts := []models.ChartOfAccount{
		{Code: "1000", Name: "الخزينة", Type: models.AccountAsset},
		{Code: "1100", Name: "المخزون", Type: models.AccountAsset},
		{Code: "1200", Name: "العملاء", Type: models.AccountAsset},
		{Code: "2000", Name: "الموردين", Type: models.AccountLiability},
		{Code: "4000", Name: "المبيعات", Type: models.AccountRevenue},
		{Code: "5000", Name: "المشتريات", Type: models.AccountExpense},
	}
	for _, account := range accounts {
		var existing models.ChartOfAccount
		if err := db.Where("code = ?", account.Code).FirstOrCreate(&existing, account).Error; err != nil {
			return err
		}
	}
	permissions := []models.Permission{
		{Code: "sales.manage", Description: "Manage sales"},
		{Code: "purchases.manage", Description: "Manage purchases"},
		{Code: "reports.view", Description: "View reports"},
		{Code: "admin.manage", Description: "Administration"},
		{Code: "delete.manage", Description: "Delete records"},
		{Code: "edit.manage", Description: "Edit records"},
		{Code: "users.manage", Description: "Manage users"},
	}
	for _, permission := range permissions {
		var existing models.Permission
		if err := db.Where("code = ?", permission.Code).FirstOrCreate(&existing, permission).Error; err != nil {
			return err
		}
	}
	var adminRole models.RBACRole
	if err := db.Where("name = ?", "admin").FirstOrCreate(&adminRole, models.RBACRole{Name: "admin", Description: "Full access"}).Error; err != nil {
		return err
	}
	var allPerms []models.Permission
	if err := db.Find(&allPerms).Error; err != nil {
		return err
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPerms); err != nil {
		return err
	}
	if err := db.Model(&admin).Association("Roles").Append(&adminRole); err != nil {
		return err
	}
	var cashierRole models.RBACRole
	if err := db.Where("name = ?", "cashier").FirstOrCreate(&cashierRole, models.RBACRole{Name: "cashier", Description: "Sales and treasury access"}).Error; err != nil {
		return err
	}
	var demoCustomer models.Customer
	if err := db.Where("name = ?", "Demo Customer").FirstOrCreate(&demoCustomer, models.Customer{Name: "Demo Customer", Phone: "01000000000", CreditLimit: 10000, BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	var demoSupplier models.Supplier
	if err := db.Where("name = ?", "Demo Supplier").FirstOrCreate(&demoSupplier, models.Supplier{Name: "Demo Supplier", Phone: "01111111111", BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	var demoItem models.Item
	if err := db.Where("code = ?", "DEMO-001").FirstOrCreate(&demoItem, models.Item{Name: "Demo Item", Code: "DEMO-001", Barcode: "622000000001", PurchasePrice: 50, SalePrice: 75, Quantity: 20, MinimumStock: 5, CategoryID: &category.ID}).Error; err != nil {
		return err
	}
	var balance models.ItemWarehouseBalance
	if err := db.Where("item_id = ? AND warehouse_id = ?", demoItem.ID, warehouse.ID).FirstOrCreate(&balance, models.ItemWarehouseBalance{ItemID: demoItem.ID, WarehouseID: warehouse.ID, Quantity: demoItem.Quantity}).Error; err != nil {
		return err
	}
	return nil
}
