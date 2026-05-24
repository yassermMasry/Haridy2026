package services

import (
	"errors"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

const TrialOperationLimit int64 = 250

const TrialLimitError = "انتهت النسخة التجريبية. برجاء الترقية للنسخة المدفوعة."
const LicenseOperationsLimitError = "انتهى عدد العمليات المسموح به في الترخيص."
const LicenseExpiredError = "انتهت مدة الترخيص. برجاء التجديد."

type UsageLimitService struct {
	db *gorm.DB
}

func NewUsageLimitService(db *gorm.DB) *UsageLimitService {
	return &UsageLimitService{db: db}
}

func (s *UsageLimitService) CheckOperationLimit(tenantID *uint) error {
	if license, ok := s.activeTenantLicense(tenantID); ok {
		if license.ExpiresAt != nil && time.Now().After(*license.ExpiresAt) {
			return errors.New(LicenseExpiredError)
		}
		if license.MaxOperations > 0 && s.CountOperations(tenantID) >= license.MaxOperations {
			return errors.New(LicenseOperationsLimitError)
		}
		return nil
	}

	limit, limited := s.operationLimit(tenantID)
	if !limited {
		return nil
	}
	if s.CountOperations(tenantID) >= limit {
		return errors.New(TrialLimitError)
	}
	return nil
}

func (s *UsageLimitService) CountOperations(tenantID *uint) int64 {
	var total int64
	total += s.countTable("sales_invoices", tenantID)
	total += s.countTable("purchase_invoices", tenantID)
	total += s.countTable("receipt_vouchers", tenantID)
	total += s.countTable("payment_vouchers", tenantID)
	return total
}

func (s *UsageLimitService) countTable(table string, tenantID *uint) int64 {
	var count int64
	query := s.db.Table(table)
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0
	}
	return count
}

func (s *UsageLimitService) operationLimit(tenantID *uint) (int64, bool) {
	if tenantID == nil {
		return TrialOperationLimit, true
	}

	var count int64
	if err := s.db.Model(&models.Subscription{}).
		Where("tenant_id = ? AND LOWER(status) = ?", *tenantID, "paid").
		Count(&count).Error; err == nil && count > 0 {
		if max := s.licenseMaxOperations(*tenantID); max > 0 {
			return max, true
		}
		return 0, false
	}

	count = 0
	if err := s.db.Model(&models.Subscription{}).
		Joins("JOIN plans ON plans.id = subscriptions.plan_id").
		Where("subscriptions.tenant_id = ? AND LOWER(subscriptions.status) = ? AND plans.price_monthly > 0", *tenantID, "active").
		Count(&count).Error; err == nil && count > 0 {
		if max := s.licenseMaxOperations(*tenantID); max > 0 {
			return max, true
		}
		return 0, false
	}
	return TrialOperationLimit, true
}

func (s *UsageLimitService) licenseMaxOperations(tenantID uint) int64 {
	var license models.LicenseKey
	if err := s.db.Where("tenant_id = ? AND used_at IS NOT NULL AND max_operations > 0", tenantID).
		Order("used_at desc").
		First(&license).Error; err != nil {
		return 0
	}
	return license.MaxOperations
}

func (s *UsageLimitService) activeTenantLicense(tenantID *uint) (models.LicenseKey, bool) {
	if tenantID == nil {
		return models.LicenseKey{}, false
	}

	var paidCount int64
	if err := s.db.Model(&models.Subscription{}).
		Where("tenant_id = ? AND LOWER(status) = ?", *tenantID, "paid").
		Count(&paidCount).Error; err != nil || paidCount == 0 {
		return models.LicenseKey{}, false
	}

	var license models.LicenseKey
	if err := s.db.Where("tenant_id = ? AND used_at IS NOT NULL", *tenantID).
		Order("used_at desc").
		First(&license).Error; err != nil {
		return models.LicenseKey{}, false
	}
	return license, true
}
