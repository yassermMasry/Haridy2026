package testutil

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"haridy2026/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	baseDBOnce sync.Once
	baseDBPath string
	baseDBErr  error
)

type Fixture struct {
	DB       *gorm.DB
	Tenant   models.Tenant
	User     models.User
	Item     models.Item
	Customer models.Customer
	Supplier models.Supplier
	Treasury models.Treasury
}

func NewFixture(t *testing.T) Fixture {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping database integration fixture in short mode")
	}
	path := filepath.Join(t.TempDir(), "haridy-test.db")
	if err := copyBaseDB(t, path); err != nil {
		t.Fatalf("copy test db: %v", err)
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap test db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	var fx Fixture
	fx.DB = db
	mustFirst(t, db.Where("slug = ?", "demo").First(&fx.Tenant), "tenant")
	mustFirst(t, db.Where("username = ?", "admin").First(&fx.User), "user")
	mustFirst(t, db.Where("code = ?", "DEMO-001").First(&fx.Item), "item")
	mustFirst(t, db.Where("name = ?", "Demo Customer").First(&fx.Customer), "customer")
	mustFirst(t, db.Where("name = ?", "Demo Supplier").First(&fx.Supplier), "supplier")
	mustFirst(t, db.Where("name = ?", "Main Treasury").First(&fx.Treasury), "treasury")
	if err := db.Model(&fx.Treasury).Update("balance", 10000).Error; err != nil {
		t.Fatalf("fund treasury: %v", err)
	}
	fx.Treasury.Balance = 10000
	return fx
}

func copyBaseDB(t *testing.T, dest string) error {
	t.Helper()
	baseDBOnce.Do(func() {
		baseDBPath, baseDBErr = createBaseDB()
	})
	if baseDBErr != nil {
		return baseDBErr
	}
	src, err := os.Open(baseDBPath)
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, src); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func createBaseDB() (string, error) {
	dir, err := os.MkdirTemp("", "haridy-base-testdb-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "base.db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return "", err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return "", err
	}
	defer sqlDB.Close()
	if err := migrateTestSchema(db); err != nil {
		return "", err
	}
	if err := seedTestData(db); err != nil {
		return "", err
	}
	return path, nil
}

func migrateTestSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Tenant{},
		&models.Branch{},
		&models.Warehouse{},
		&models.User{},
		&models.Customer{},
		&models.CustomerTransaction{},
		&models.Supplier{},
		&models.SupplierTransaction{},
		&models.ItemCategory{},
		&models.Item{},
		&models.StockMovement{},
		&models.SalesInvoice{},
		&models.SalesInvoiceItem{},
		&models.PurchaseInvoice{},
		&models.PurchaseInvoiceItem{},
		&models.Treasury{},
		&models.TreasuryTransaction{},
		&models.ChartOfAccount{},
		&models.JournalEntry{},
		&models.JournalEntryLine{},
		&models.AuditLog{},
		&models.SecurityEvent{},
		&models.LoginAttempt{},
		&models.SalesReturn{},
		&models.SalesReturnItem{},
		&models.PurchaseReturn{},
		&models.PurchaseReturnItem{},
	)
}

func seedTestData(db *gorm.DB) error {
	tenant := models.Tenant{Name: "Demo Company", Slug: "demo", Subdomain: "demo", Status: "trial"}
	if err := db.Create(&tenant).Error; err != nil {
		return err
	}
	branch := models.Branch{TenantID: &tenant.ID, Name: "Main Branch", Code: "MAIN", IsActive: true}
	if err := db.Create(&branch).Error; err != nil {
		return err
	}
	warehouse := models.Warehouse{TenantID: &tenant.ID, Name: "Main Warehouse", Code: "MAIN-WH", BranchID: branch.ID, IsActive: true}
	if err := db.Create(&warehouse).Error; err != nil {
		return err
	}
	user := models.User{TenantID: &tenant.ID, Username: "admin", PasswordHash: "test", Role: models.RoleAdmin, CurrentBranchID: &branch.ID}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	category := models.ItemCategory{TenantID: &tenant.ID, Name: "General"}
	if err := db.Create(&category).Error; err != nil {
		return err
	}
	item := models.Item{TenantID: &tenant.ID, Name: "Demo Item", Code: "DEMO-001", Barcode: "622000000001", PurchasePrice: 50, SalePrice: 75, Quantity: 20, MinimumStock: 5, CategoryID: &category.ID}
	if err := db.Create(&item).Error; err != nil {
		return err
	}
	if err := db.Create(&models.Customer{TenantID: &tenant.ID, Name: "Demo Customer", Phone: "01000000000", CreditLimit: 10000, BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	if err := db.Create(&models.Supplier{TenantID: &tenant.ID, Name: "Demo Supplier", Phone: "01111111111", BranchID: &branch.ID}).Error; err != nil {
		return err
	}
	if err := db.Create(&models.Treasury{TenantID: &tenant.ID, BranchID: &branch.ID, Name: "Main Treasury", Balance: 10000}).Error; err != nil {
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
	return db.Create(&accounts).Error
}

func ChdirRepoRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				t.Fatalf("change to repo root: %v", err)
			}
			t.Cleanup(func() { _ = os.Chdir(wd) })
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func mustFirst(t *testing.T, tx *gorm.DB, label string) {
	t.Helper()
	if tx.Error != nil {
		t.Fatalf("load %s: %v", label, tx.Error)
	}
}
