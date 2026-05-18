package repositories

import (
	"testing"

	"haridy2026/internal/models"
	"haridy2026/internal/testutil"
)

func TestRepositoryUsesSeededDatabase(t *testing.T) {
	fx := testutil.NewFixture(t)
	repo := New(fx.DB)
	var count int64
	if err := repo.DB.Model(&models.Item{}).Where("tenant_id = ?", fx.Tenant.ID).Count(&count).Error; err != nil {
		t.Fatalf("count tenant items: %v", err)
	}
	if count == 0 {
		t.Fatal("expected seeded tenant items")
	}
}
