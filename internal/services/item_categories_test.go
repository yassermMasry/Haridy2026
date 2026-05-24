package services

import (
	"testing"

	"haridy2026/internal/models"
	"haridy2026/internal/testutil"
)

func TestCreateItemCategory(t *testing.T) {
	fx := testutil.NewFixture(t)
	category, err := NewItemCategoryService(fx.DB).Create(fx.Tenant.ID, "منظفات")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if category.ID == 0 || category.Name != "منظفات" || category.TenantID == nil || *category.TenantID != fx.Tenant.ID {
		t.Fatalf("unexpected category: %+v", category)
	}
}

func TestDuplicateItemCategoryFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemCategoryService(fx.DB)
	if _, err := service.Create(fx.Tenant.ID, "منظفات"); err != nil {
		t.Fatalf("create category: %v", err)
	}
	if _, err := service.Create(fx.Tenant.ID, "منظفات"); err == nil {
		t.Fatal("expected duplicate category to fail")
	}
}

func TestUpdateUsedItemCategorySucceeds(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemCategoryService(fx.DB)
	category, err := service.Create(fx.Tenant.ID, "قابل للتعديل")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := fx.DB.Model(&models.Item{}).Where("id = ?", fx.Item.ID).Update("category_id", category.ID).Error; err != nil {
		t.Fatalf("link item: %v", err)
	}
	if err := service.Update(fx.Tenant.ID, category.ID, "تم التعديل"); err != nil {
		t.Fatalf("update used category: %v", err)
	}
}

func TestDeleteUsedItemCategoryFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	categoryID := *fx.Item.CategoryID
	err := NewItemCategoryService(fx.DB).Delete(categoryID)
	if err == nil {
		t.Fatal("expected delete used category to fail")
	}
	if err.Error() != CategoryInUseMessage {
		t.Fatalf("unexpected error: %v", err)
	}
}
