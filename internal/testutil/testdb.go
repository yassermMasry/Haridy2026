package testutil

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"haridy2026/internal/database"
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
	if err := database.AutoMigrate(db); err != nil {
		return "", err
	}
	if err := database.Seed(db); err != nil {
		return "", err
	}
	return path, nil
}

func mustFirst(t *testing.T, tx *gorm.DB, label string) {
	t.Helper()
	if tx.Error != nil {
		t.Fatalf("load %s: %v", label, tx.Error)
	}
}
