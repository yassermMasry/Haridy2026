package database

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"haridy2026/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seed(db *gorm.DB) error {
	var plan models.Plan
	if err := db.Where("code = ?", "starter").FirstOrCreate(&plan, models.Plan{Code: "starter", Name: "Starter", PriceMonthly: 49, MaxUsers: 5, MaxBranches: 1, Features: "sales,purchases,inventory,treasury", IsActive: true}).Error; err != nil {
		return err
	}
	var proPlan models.Plan
	if err := db.Where("code = ?", "pro").FirstOrCreate(&proPlan, models.Plan{Code: "pro", Name: "Professional", PriceMonthly: 149, MaxUsers: 25, MaxBranches: 5, Features: "all", IsActive: true}).Error; err != nil {
		return err
	}
	var tenant models.Tenant
	if err := db.Where("slug = ?", "demo").FirstOrCreate(&tenant, models.Tenant{Name: "Demo Company", Slug: "demo", Subdomain: "demo", Status: "trial"}).Error; err != nil {
		return err
	}
	var setting models.CompanySetting
	if err := db.Where("tenant_id = ?", tenant.ID).FirstOrCreate(&setting, models.CompanySetting{TenantID: tenant.ID, LegalName: "Demo Company LLC", TaxNumber: "VAT-DEMO-001", Currency: "EGP", TimeZone: "Africa/Cairo"}).Error; err != nil {
		return err
	}
	var sub models.Subscription
	if err := db.Where("tenant_id = ?", tenant.ID).FirstOrCreate(&sub, models.Subscription{TenantID: tenant.ID, PlanID: plan.ID, Status: "trial", StartsAt: time.Now()}).Error; err != nil {
		return err
	}
	var branch models.Branch
	if err := db.Where("code = ?", "MAIN").FirstOrCreate(&branch, models.Branch{TenantID: &tenant.ID, Name: "Main Branch", Code: "MAIN", IsActive: true}).Error; err != nil {
		return err
	}
	var warehouse models.Warehouse
	if err := db.Where("code = ?", "MAIN-WH").FirstOrCreate(&warehouse, models.Warehouse{TenantID: &tenant.ID, Name: "Main Warehouse", Code: "MAIN-WH", BranchID: branch.ID, IsActive: true}).Error; err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var admin models.User
	if err := db.Unscoped().Where("username = ?", "admin").First(&admin).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		admin = models.User{
			TenantID:        &tenant.ID,
			Username:        "admin",
			PasswordHash:    string(hash),
			Role:            models.RoleAdmin,
			CurrentBranchID: &branch.ID,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	} else {
		admin.TenantID = &tenant.ID
		admin.Role = models.RoleAdmin
		admin.CurrentBranchID = &branch.ID
		admin.PasswordHash = string(hash)
		admin.DeletedAt = gorm.DeletedAt{}
		if err := db.Unscoped().Save(&admin).Error; err != nil {
			return err
		}
	}
	var tenantUser models.TenantUser
	if err := db.Where("tenant_id = ? AND user_id = ?", tenant.ID, admin.ID).FirstOrCreate(&tenantUser, models.TenantUser{TenantID: tenant.ID, UserID: admin.ID, IsOwner: true}).Error; err != nil {
		return err
	}

	var treasury models.Treasury
	if err := db.Where("name = ?", "Main Treasury").FirstOrCreate(&treasury, models.Treasury{TenantID: &tenant.ID, BranchID: &branch.ID, Name: "Main Treasury"}).Error; err != nil {
		return err
	}
	var category models.ItemCategory
	if err := db.Where("name = ?", "General").FirstOrCreate(&category, models.ItemCategory{TenantID: &tenant.ID, Name: "General"}).Error; err != nil {
		return err
	}
	accounts := []models.ChartOfAccount{
		{TenantID: &tenant.ID, Code: "1000", Name: "Cash", Type: models.AccountAsset},
		{TenantID: &tenant.ID, Code: "1100", Name: "Inventory", Type: models.AccountAsset},
		{TenantID: &tenant.ID, Code: "1200", Name: "Accounts Receivable", Type: models.AccountAsset},
		{TenantID: &tenant.ID, Code: "2000", Name: "Accounts Payable", Type: models.AccountLiability},
		{TenantID: &tenant.ID, Code: "4000", Name: "Sales", Type: models.AccountRevenue},
		{TenantID: &tenant.ID, Code: "5000", Name: "Purchases", Type: models.AccountExpense},
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
	var adminRoleCount int64
	if err := db.Model(&models.UserRole{}).
		Where("user_id = ? AND rbac_role_id = ?", admin.ID, adminRole.ID).
		Count(&adminRoleCount).Error; err != nil {
		return err
	}
	if adminRoleCount == 0 {
		if err := db.Model(&admin).Association("Roles").Append(&adminRole); err != nil {
			return err
		}
	}
	var cashierRole models.RBACRole
	if err := db.Where("name = ?", "cashier").FirstOrCreate(&cashierRole, models.RBACRole{Name: "cashier", Description: "Sales and treasury access"}).Error; err != nil {
		return err
	}
	var demoCustomer models.Customer
	if err := db.Where("name = ?", "Demo Customer").FirstOrCreate(&demoCustomer, models.Customer{TenantID: &tenant.ID, Name: "Demo Customer", Phone: "01000000000", CreditLimit: 10000, BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	var demoSupplier models.Supplier
	if err := db.Where("name = ?", "Demo Supplier").FirstOrCreate(&demoSupplier, models.Supplier{TenantID: &tenant.ID, Name: "Demo Supplier", Phone: "01111111111", BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	var demoItem models.Item
	if err := db.Where("code = ?", "DEMO-001").FirstOrCreate(&demoItem, models.Item{TenantID: &tenant.ID, Name: "Demo Item", Code: "DEMO-001", Barcode: "622000000001", PurchasePrice: 50, SalePrice: 75, Quantity: 20, MinimumStock: 5, CategoryID: &category.ID}).Error; err != nil {
		return err
	}
	var balance models.ItemWarehouseBalance
	if err := db.Where("item_id = ? AND warehouse_id = ?", demoItem.ID, warehouse.ID).FirstOrCreate(&balance, models.ItemWarehouseBalance{ItemID: demoItem.ID, WarehouseID: warehouse.ID, Quantity: demoItem.Quantity}).Error; err != nil {
		return err
	}
	var pipeline models.SalesPipeline
	if err := db.Where("tenant_id = ? AND name = ?", tenant.ID, "Default Pipeline").FirstOrCreate(&pipeline, models.SalesPipeline{TenantID: tenant.ID, Name: "Default Pipeline"}).Error; err != nil {
		return err
	}
	for i, stage := range []string{"Lead", "Qualified", "Proposal", "Won"} {
		var existing models.SalesPipelineStage
		if err := db.Where("pipeline_id = ? AND name = ?", pipeline.ID, stage).FirstOrCreate(&existing, models.SalesPipelineStage{PipelineID: pipeline.ID, Name: stage, SortOrder: i + 1}).Error; err != nil {
			return err
		}
	}
	var fy models.FiscalYear
	now := time.Now()
	if err := db.Where("tenant_id = ? AND name = ?", tenant.ID, now.Format("2006")).FirstOrCreate(&fy, models.FiscalYear{TenantID: tenant.ID, Name: now.Format("2006"), StartDate: time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.UTC)}).Error; err != nil {
		return err
	}
	demoLicenseHash := hashSeedLicenseCode("DEMO-LICENSE")
	var licenseCount int64
	if err := db.Model(&models.LicenseKey{}).
		Where("key_hash = ? OR (tenant_id = ? AND plan_code = ? AND status = ?)", demoLicenseHash, tenant.ID, "starter", "used").
		Count(&licenseCount).Error; err != nil {
		return err
	}
	if licenseCount == 0 {
		license := models.LicenseKey{TenantID: &tenant.ID, Key: demoLicenseHash, KeyHash: demoLicenseHash, PlanCode: "starter", MaxOperations: 100000, Status: "used"}
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key_hash"}}, DoNothing: true}).Create(&license).Error; err != nil {
			return err
		}
	}
	return nil
}

func hashSeedLicenseCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
