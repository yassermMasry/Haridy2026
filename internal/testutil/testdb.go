package testutil

import (
	"path/filepath"
	"testing"

	"haridy2026/internal/database"
	"haridy2026/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "haridy-test.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap test db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed test db: %v", err)
	}
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

func mustFirst(t *testing.T, tx *gorm.DB, label string) {
	t.Helper()
	if tx.Error != nil {
		t.Fatalf("load %s: %v", label, tx.Error)
	}
}
