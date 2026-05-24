package database

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAutoMigrateBackfillsLegacyLicenseKeys(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/legacy.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.Exec(`CREATE TABLE license_keys (
		id integer primary key autoincrement,
		key text,
		created_at datetime
	)`).Error; err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if err := db.Exec(`INSERT INTO license_keys (key) VALUES (?), (?)`, "ABC-123", "").Error; err != nil {
		t.Fatalf("insert legacy rows: %v", err)
	}

	if err := AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	var activeHash string
	if err := db.Raw(`SELECT key_hash FROM license_keys WHERE key = ?`, "ABC-123").Scan(&activeHash).Error; err != nil {
		t.Fatalf("load migrated hash: %v", err)
	}
	if activeHash != legacyLicenseHash("ABC-123") {
		t.Fatalf("unexpected key_hash: %q", activeHash)
	}

	var invalid struct {
		KeyHash string
		Status  string
	}
	if err := db.Raw(`SELECT key_hash, status FROM license_keys WHERE key = ''`).Scan(&invalid).Error; err != nil {
		t.Fatalf("load invalid row: %v", err)
	}
	if invalid.KeyHash == "" || invalid.Status != "legacy_invalid" {
		t.Fatalf("expected invalid legacy row to be marked safely, got %+v", invalid)
	}
}
