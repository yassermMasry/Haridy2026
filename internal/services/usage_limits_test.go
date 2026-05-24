package services

import (
	"fmt"
	"testing"
	"time"

	"haridy2026/internal/models"
	"haridy2026/internal/testutil"
)

func TestTrialLimitAllowsBefore250(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 249); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}

	limits := NewUsageLimitService(fx.DB)
	if got := limits.CountOperations(&tenantID); got != 249 {
		t.Fatalf("expected 249 operations, got %d", got)
	}
	if err := limits.CheckOperationLimit(&tenantID); err != nil {
		t.Fatalf("expected trial limit to allow operation: %v", err)
	}
}

func TestTrialLimitBlocksAt250(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 250); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}

	err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID)
	if err == nil {
		t.Fatal("expected trial limit error")
	}
	if err.Error() != TrialLimitError {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaidSubscriptionDoesNotBlockAtLimit(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 250); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}
	if err := fx.DB.AutoMigrate(&models.Plan{}, &models.Subscription{}); err != nil {
		t.Fatalf("migrate subscription tables: %v", err)
	}
	plan := models.Plan{Code: "paid-test", Name: "Paid Test", PriceMonthly: 100, IsActive: true}
	if err := fx.DB.Create(&plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if err := fx.DB.Create(&models.Subscription{TenantID: tenantID, PlanID: plan.ID, Status: "paid"}).Error; err != nil {
		t.Fatalf("create paid subscription: %v", err)
	}

	if err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID); err != nil {
		t.Fatalf("expected paid subscription to bypass limit: %v", err)
	}
}

func TestPaidLicenseWithinDurationAllows(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 250); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}
	createPaidLicense(t, fx, tenantID, 251, time.Now().AddDate(0, 0, 30))

	if err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID); err != nil {
		t.Fatalf("expected paid license within duration to allow operation: %v", err)
	}
}

func TestPaidLicenseExpiredBlocks(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	createPaidLicense(t, fx, tenantID, 100000, time.Now().AddDate(0, 0, -1))

	err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID)
	if err == nil {
		t.Fatal("expected expired license to block operation")
	}
	if err.Error() != LicenseExpiredError {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaidLicenseOperationLimitBlocks(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 250); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}
	createPaidLicense(t, fx, tenantID, 250, time.Now().AddDate(0, 0, 30))

	err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID)
	if err == nil {
		t.Fatal("expected paid license operation limit to block operation")
	}
	if err.Error() != LicenseOperationsLimitError {
		t.Fatalf("unexpected error: %v", err)
	}
}

func seedSalesInvoices(fx testutil.Fixture, tenantID uint, count int) error {
	invoices := make([]models.SalesInvoice, 0, count)
	for i := 0; i < count; i++ {
		invoices = append(invoices, models.SalesInvoice{
			TenantID:    &tenantID,
			Number:      fmt.Sprintf("LIMIT-SALE-%03d", i),
			UserID:      fx.User.ID,
			PaymentType: "cash",
			Total:       1,
		})
	}
	return fx.DB.Create(&invoices).Error
}

func createPaidLicense(t *testing.T, fx testutil.Fixture, tenantID uint, maxOperations int64, expiresAt time.Time) {
	t.Helper()
	if err := fx.DB.AutoMigrate(&models.Plan{}, &models.Subscription{}, &models.LicenseKey{}); err != nil {
		t.Fatalf("migrate license tables: %v", err)
	}
	plan := models.Plan{Code: fmt.Sprintf("paid-%d", time.Now().UnixNano()), Name: "Paid", PriceMonthly: 100, IsActive: true}
	if err := fx.DB.Create(&plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	if err := fx.DB.Create(&models.Subscription{TenantID: tenantID, PlanID: plan.ID, Status: "paid", StartsAt: time.Now(), EndsAt: &expiresAt}).Error; err != nil {
		t.Fatalf("create paid subscription: %v", err)
	}
	usedAt := time.Now()
	hash := HashLicenseCode(fmt.Sprintf("TEST-LICENSE-%d", usedAt.UnixNano()))
	if err := fx.DB.Create(&models.LicenseKey{
		TenantID:      &tenantID,
		Key:           hash,
		KeyHash:       hash,
		PlanCode:      plan.Code,
		MaxOperations: maxOperations,
		Status:        "used",
		ExpiresAt:     &expiresAt,
		UsedAt:        &usedAt,
	}).Error; err != nil {
		t.Fatalf("create license: %v", err)
	}
}
