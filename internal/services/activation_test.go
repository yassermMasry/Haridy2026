package services

import (
	"testing"
	"time"

	"haridy2026/internal/models"
	"haridy2026/internal/testutil"
)

func TestActivationSucceedsAndRaisesOperationLimit(t *testing.T) {
	fx := testutil.NewFixture(t)
	prepareActivationTables(t, fx)
	tenantID := fx.Tenant.ID
	if err := seedSalesInvoices(fx, tenantID, 250); err != nil {
		t.Fatalf("seed sales invoices: %v", err)
	}

	service := NewActivationService(fx.DB)
	expiresAt := time.Now().AddDate(0, 0, 365)
	code, _, err := service.CreateLicense("yearly", 100000, &expiresAt)
	if err != nil {
		t.Fatalf("create license: %v", err)
	}
	if err := service.Activate(tenantID, code); err != nil {
		t.Fatalf("activate license: %v", err)
	}

	var license models.LicenseKey
	if err := fx.DB.Where("key_hash = ?", HashLicenseCode(code)).First(&license).Error; err != nil {
		t.Fatalf("load license: %v", err)
	}
	if license.TenantID == nil || *license.TenantID != tenantID || license.UsedAt == nil || license.Status != "used" {
		t.Fatalf("license was not marked as used for tenant: %+v", license)
	}
	var sub models.Subscription
	if err := fx.DB.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		t.Fatalf("load subscription: %v", err)
	}
	if sub.Status != "paid" {
		t.Fatalf("expected paid subscription, got %q", sub.Status)
	}
	if sub.EndsAt == nil {
		t.Fatal("expected subscription expiry to be stored")
	}
	if err := NewUsageLimitService(fx.DB).CheckOperationLimit(&tenantID); err != nil {
		t.Fatalf("expected raised operation limit: %v", err)
	}
}

func TestActivationRejectsWrongCode(t *testing.T) {
	fx := testutil.NewFixture(t)
	prepareActivationTables(t, fx)

	err := NewActivationService(fx.DB).Activate(fx.Tenant.ID, "WRONG-CODE")
	if err == nil {
		t.Fatal("expected wrong code to fail")
	}
}

func TestActivationRejectsUsedCode(t *testing.T) {
	fx := testutil.NewFixture(t)
	prepareActivationTables(t, fx)

	service := NewActivationService(fx.DB)
	code, _, err := service.CreateLicense("yearly", 100000, nil)
	if err != nil {
		t.Fatalf("create license: %v", err)
	}
	if err := service.Activate(fx.Tenant.ID, code); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if err := service.Activate(fx.Tenant.ID, code); err == nil {
		t.Fatal("expected used code to fail")
	}
}

func prepareActivationTables(t *testing.T, fx testutil.Fixture) {
	t.Helper()
	if err := fx.DB.AutoMigrate(&models.Plan{}, &models.Subscription{}, &models.LicenseKey{}); err != nil {
		t.Fatalf("migrate activation tables: %v", err)
	}
}
