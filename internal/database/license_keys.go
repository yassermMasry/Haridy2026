package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

func migrateLicenseKeys(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.LicenseKey{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&models.LicenseKey{}, "key_hash") {
		if err := db.Migrator().AddColumn(&models.LicenseKey{}, "KeyHash"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&models.LicenseKey{}, "status") {
		if err := db.Migrator().AddColumn(&models.LicenseKey{}, "Status"); err != nil {
			return err
		}
	}

	var rows []models.LicenseKey
	if err := db.
		Select("id", "key", "key_hash", "status").
		Where("(key_hash IS NULL OR key_hash = '')").
		Find(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		updates := map[string]any{}
		if hash := legacyLicenseHash(row.Key); hash != "" {
			updates["key_hash"] = hash
		} else {
			updates["key_hash"] = fmt.Sprintf("legacy-invalid-%d", row.ID)
			updates["status"] = "legacy_invalid"
		}
		if err := db.Model(&models.LicenseKey{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	if db.Dialector.Name() == "postgres" {
		if err := db.Exec("ALTER TABLE license_keys ALTER COLUMN key_hash SET NOT NULL").Error; err != nil {
			return err
		}
	}
	return nil
}

func legacyLicenseHash(key string) string {
	normalized := strings.ToUpper(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	if normalized == "" {
		return ""
	}
	if len(normalized) == 64 && isHex(normalized) {
		return strings.ToLower(normalized)
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func isHex(value string) bool {
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
