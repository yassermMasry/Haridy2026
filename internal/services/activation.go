package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ActivationSuccessMessage = "تم تفعيل النسخة بنجاح"

type ActivationService struct {
	db *gorm.DB
}

func NewActivationService(db *gorm.DB) *ActivationService {
	return &ActivationService{db: db}
}

func (s *ActivationService) Activate(tenantID uint, code string) error {
	if tenantID == 0 {
		return errors.New("tenant is required")
	}
	hash := HashLicenseCode(code)
	if hash == "" {
		return errors.New("invalid activation code")
	}

	now := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		var license models.LicenseKey
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("key_hash = ? AND status = ? AND used_at IS NULL AND tenant_id IS NULL", hash, "active").
			First(&license).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid activation code")
			}
			return err
		}
		if license.ExpiresAt != nil && license.ExpiresAt.Before(now) {
			return errors.New("activation code expired")
		}

		updates := map[string]any{
			"tenant_id": tenantID,
			"used_at":   now,
			"status":    "used",
		}
		if err := tx.Model(&license).Updates(updates).Error; err != nil {
			return err
		}

		planCode := strings.TrimSpace(license.PlanCode)
		if planCode == "" {
			planCode = "paid"
		}
		plan := models.Plan{Code: planCode, Name: planCode, PriceMonthly: 1, IsActive: true}
		if err := tx.Where("code = ?", planCode).FirstOrCreate(&plan).Error; err != nil {
			return err
		}

		var sub models.Subscription
		err := tx.Where("tenant_id = ?", tenantID).First(&sub).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sub = models.Subscription{TenantID: tenantID, PlanID: plan.ID, Status: "paid", StartsAt: now, EndsAt: license.ExpiresAt}
			return tx.Create(&sub).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&sub).Updates(map[string]any{
			"plan_id":   plan.ID,
			"status":    "paid",
			"starts_at": now,
			"ends_at":   license.ExpiresAt,
		}).Error
	})
}

func (s *ActivationService) CreateLicense(planCode string, maxOperations int64, expiresAt *time.Time) (string, *models.LicenseKey, error) {
	code, err := GenerateLicenseCode()
	if err != nil {
		return "", nil, err
	}
	hash := HashLicenseCode(code)
	license := models.LicenseKey{
		Key:           hash,
		KeyHash:       hash,
		PlanCode:      strings.TrimSpace(planCode),
		MaxOperations: maxOperations,
		ExpiresAt:     expiresAt,
		Status:        "active",
	}
	if license.PlanCode == "" {
		license.PlanCode = "yearly"
	}
	if err := s.db.Create(&license).Error; err != nil {
		return "", nil, err
	}
	return code, &license, nil
}

func GenerateLicenseCode() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	encoded := strings.ToUpper(hex.EncodeToString(bytes))
	return fmt.Sprintf("%s-%s-%s-%s", encoded[0:6], encoded[6:12], encoded[12:18], encoded[18:24]), nil
}

func HashLicenseCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
