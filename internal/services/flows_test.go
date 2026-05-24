package services

import (
	"testing"

	"haridy2026/internal/models"
	"haridy2026/internal/testutil"
)

func TestSalesPurchaseReturnsTreasuryAndAccountingConsistency(t *testing.T) {
	fx := testutil.NewFixture(t)
	sales := NewSalesService(fx.DB)
	purchases := NewPurchaseService(fx.DB)
	returns := NewReturnService(fx.DB)
	treasury := NewTreasuryService(fx.DB)

	sale, err := sales.Create(SaleInput{
		UserID:      fx.User.ID,
		CustomerID:  fx.Customer.ID,
		PaymentType: "cash",
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 2, UnitPrice: 80}},
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}
	purchase, err := purchases.Create(PurchaseInput{
		UserID:      fx.User.ID,
		SupplierID:  fx.Supplier.ID,
		PaymentType: "cash",
		Lines:       []PurchaseLineInput{{ItemID: fx.Item.ID, Quantity: 3, UnitCost: 40}},
	})
	if err != nil {
		t.Fatalf("create purchase: %v", err)
	}
	if err := returns.SalesReturn(sale.ID, "test", fx.User.ID); err != nil {
		t.Fatalf("sales return: %v", err)
	}
	if err := returns.PurchaseReturn(purchase.ID, "test", fx.User.ID); err != nil {
		t.Fatalf("purchase return: %v", err)
	}
	if err := treasury.Add(models.TreasuryReceive, 25, "test receive", fx.User.ID); err != nil {
		t.Fatalf("treasury receive: %v", err)
	}
	if findings, err := ReconcileAccounting(fx.DB); err != nil || findings != "" {
		t.Fatalf("accounting reconciliation failed: findings=%q err=%v", findings, err)
	}
	var item models.Item
	if err := fx.DB.First(&item, fx.Item.ID).Error; err != nil {
		t.Fatalf("reload item: %v", err)
	}
	if item.Quantity != fx.Item.Quantity {
		t.Fatalf("expected inventory to return to %.3f, got %.3f", fx.Item.Quantity, item.Quantity)
	}
}

func TestNegativeInventoryIsBlocked(t *testing.T) {
	fx := testutil.NewFixture(t)
	_, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		PaymentType: "cash",
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: fx.Item.Quantity + 1, UnitPrice: 80}},
	})
	if err == nil {
		t.Fatal("expected sale that exceeds stock to fail")
	}
}

func TestDuplicatePostingIsBlocked(t *testing.T) {
	fx := testutil.NewFixture(t)
	lines := []JournalLineInput{{AccountCode: "1000", Debit: 10}, {AccountCode: "4000", Credit: 10}}
	if err := CreateJournal(fx.DB, "DUP-1", "first", lines); err != nil {
		t.Fatalf("first posting: %v", err)
	}
	if err := CreateJournal(fx.DB, "DUP-1", "second", lines); err == nil {
		t.Fatal("expected duplicate posting to fail")
	}
}

func TestCreateWarehouseValidatesAndSaves(t *testing.T) {
	fx := testutil.NewFixture(t)
	erp := NewERPService(fx.DB)
	if fx.Treasury.BranchID == nil {
		t.Fatal("fixture treasury has no branch")
	}
	if err := erp.CreateWarehouse(fx.Tenant.ID, *fx.Treasury.BranchID, "Secondary Warehouse", "SEC-WH"); err != nil {
		t.Fatalf("create warehouse: %v", err)
	}
	var warehouse models.Warehouse
	if err := fx.DB.Where("code = ?", "SEC-WH").First(&warehouse).Error; err != nil {
		t.Fatalf("load warehouse: %v", err)
	}
	if warehouse.Name != "Secondary Warehouse" || warehouse.BranchID == 0 {
		t.Fatalf("unexpected warehouse: %+v", warehouse)
	}
	if err := erp.CreateWarehouse(fx.Tenant.ID, warehouse.BranchID, "Duplicate", "SEC-WH"); err == nil {
		t.Fatal("expected duplicate warehouse code to fail")
	}
}

func TestPriceListAndUpdateSalePrice(t *testing.T) {
	fx := testutil.NewFixture(t)
	items := NewItemService(fx.DB)
	if _, err := items.UpdateSalePrice(fx.Item.ID, -1); err == nil {
		t.Fatal("expected negative sale price to fail")
	}
	warn, err := items.UpdateSalePrice(fx.Item.ID, fx.Item.PurchasePrice-1)
	if err != nil {
		t.Fatalf("update sale price: %v", err)
	}
	if !warn {
		t.Fatal("expected below-cost warning")
	}
	list := items.PriceList("DEMO-001", "loss")
	if len(list.Rows) != 1 {
		t.Fatalf("expected one loss item, got %d", len(list.Rows))
	}
	if list.Rows[0].Status != PriceStatusLoss {
		t.Fatalf("expected loss status, got %s", list.Rows[0].Status)
	}
}

func TestDeleteUnusedItemSucceeds(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemService(fx.DB)
	item := models.Item{Name: "Unused Item", Code: "UNUSED-001", Barcode: "UNUSED-BAR", PurchasePrice: 5, SalePrice: 7}
	if err := service.Save(&item); err != nil {
		t.Fatalf("create unused item: %v", err)
	}
	if err := service.Delete(item.ID); err != nil {
		t.Fatalf("delete unused item: %v", err)
	}
	var count int64
	if err := fx.DB.Model(&models.Item{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
		t.Fatalf("count item: %v", err)
	}
	if count != 0 {
		t.Fatal("expected deleted item to be hidden by soft delete")
	}
}

func TestDeleteUsedItemFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemService(fx.DB)
	if err := fx.DB.Create(&models.StockMovement{TenantID: &fx.Tenant.ID, ItemID: fx.Item.ID, Type: models.StockIn, Quantity: 1}).Error; err != nil {
		t.Fatalf("create stock movement: %v", err)
	}
	err := service.Delete(fx.Item.ID)
	if err == nil {
		t.Fatal("expected used item delete to fail")
	}
	if err.Error() != "لا يمكن حذف الصنف لأنه مستخدم في عمليات سابقة" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateItemDuplicateCodeFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemService(fx.DB)
	item := models.Item{Name: "Other Item", Code: "OTHER-DUP", Barcode: "OTHER-DUP-BAR", PurchasePrice: 5, SalePrice: 7}
	if err := service.Save(&item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	item.Code = fx.Item.Code
	if err := service.Save(&item); err == nil {
		t.Fatal("expected duplicate code update to fail")
	}
}

func TestUpdateItemEmptyNameFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemService(fx.DB)
	item := fx.Item
	item.Name = ""
	if err := service.Save(&item); err == nil {
		t.Fatal("expected empty name update to fail")
	}
}
