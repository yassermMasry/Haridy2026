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

func TestCashSaleFullyPaidWithoutCustomer(t *testing.T) {
	fx := testutil.NewFixture(t)
	invoice, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		PaymentType: "cash",
		PaidCash:    20,
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 1, UnitPrice: 20}},
	})
	if err != nil {
		t.Fatalf("create cash sale: %v", err)
	}
	if invoice.Total != 20 || invoice.PaidCash != 20 || invoice.CustomerID != nil {
		t.Fatalf("unexpected invoice: %+v", invoice)
	}
	var treasuryTx models.TreasuryTransaction
	if err := fx.DB.Where("reference = ? AND amount = ?", invoice.Number, 20).First(&treasuryTx).Error; err != nil {
		t.Fatalf("load treasury tx: %v", err)
	}
}

func TestPartialPaidSalePostsRemainingToCustomer(t *testing.T) {
	fx := testutil.NewFixture(t)
	invoice, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		CustomerID:  fx.Customer.ID,
		PaymentType: "cash",
		PaidCash:    10,
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 1, UnitPrice: 25}},
	})
	if err != nil {
		t.Fatalf("create partial sale: %v", err)
	}
	var tx models.CustomerTransaction
	if err := fx.DB.Where("customer_id = ? AND reference = ?", fx.Customer.ID, invoice.Number).First(&tx).Error; err != nil {
		t.Fatalf("load customer tx: %v", err)
	}
	if tx.Debit != 15 {
		t.Fatalf("expected remaining 15, got %.2f", tx.Debit)
	}
}

func TestOverpaidSaleCreatesCustomerCredit(t *testing.T) {
	fx := testutil.NewFixture(t)
	invoice, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		CustomerID:  fx.Customer.ID,
		PaymentType: "cash",
		PaidCash:    30,
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 1, UnitPrice: 20}},
	})
	if err != nil {
		t.Fatalf("create overpaid sale: %v", err)
	}
	var tx models.CustomerTransaction
	if err := fx.DB.Where("customer_id = ? AND reference = ?", fx.Customer.ID, invoice.Number).First(&tx).Error; err != nil {
		t.Fatalf("load customer tx: %v", err)
	}
	if tx.Credit != 10 {
		t.Fatalf("expected credit 10, got %.2f", tx.Credit)
	}
}

func TestSaleFromSpecificWarehouseDeductsOnlyThatWarehouse(t *testing.T) {
	fx := testutil.NewFixture(t)
	var warehouse models.Warehouse
	if err := fx.DB.Where("code = ?", "MAIN-WH").First(&warehouse).Error; err != nil {
		t.Fatalf("load warehouse: %v", err)
	}
	if err := fx.DB.Create(&models.ItemWarehouseBalance{ItemID: fx.Item.ID, WarehouseID: warehouse.ID, Quantity: 5}).Error; err != nil {
		t.Fatalf("create warehouse balance: %v", err)
	}
	if _, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		WarehouseID: warehouse.ID,
		PaymentType: "cash",
		PaidCash:    20,
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 2, UnitPrice: 10}},
	}); err != nil {
		t.Fatalf("create warehouse sale: %v", err)
	}
	var balance models.ItemWarehouseBalance
	if err := fx.DB.Where("item_id = ? AND warehouse_id = ?", fx.Item.ID, warehouse.ID).First(&balance).Error; err != nil {
		t.Fatalf("load balance: %v", err)
	}
	if balance.Quantity != 3 {
		t.Fatalf("expected balance 3, got %.3f", balance.Quantity)
	}
}

func TestSaleFromSpecificWarehouseBlocksUnavailableQuantity(t *testing.T) {
	fx := testutil.NewFixture(t)
	var warehouse models.Warehouse
	if err := fx.DB.Where("code = ?", "MAIN-WH").First(&warehouse).Error; err != nil {
		t.Fatalf("load warehouse: %v", err)
	}
	if err := fx.DB.Create(&models.ItemWarehouseBalance{ItemID: fx.Item.ID, WarehouseID: warehouse.ID, Quantity: 1}).Error; err != nil {
		t.Fatalf("create warehouse balance: %v", err)
	}
	_, err := NewSalesService(fx.DB).Create(SaleInput{
		UserID:      fx.User.ID,
		WarehouseID: warehouse.ID,
		PaymentType: "cash",
		PaidCash:    20,
		Lines:       []SaleLineInput{{ItemID: fx.Item.ID, Quantity: 2, UnitPrice: 10}},
	})
	if err == nil {
		t.Fatal("expected unavailable warehouse quantity to fail")
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

func TestPriceListAllUsesWarehouseTotalQuantity(t *testing.T) {
	fx := testutil.NewFixture(t)
	list := NewItemService(fx.DB).PriceList("DEMO-001", "all")
	if len(list.Rows) != 1 {
		t.Fatalf("expected one item, got %d", len(list.Rows))
	}
	if list.Rows[0].Quantity != fx.Item.Quantity {
		t.Fatalf("expected total quantity %.3f, got %.3f", fx.Item.Quantity, list.Rows[0].Quantity)
	}
	if list.Rows[0].WarehouseName != "" {
		t.Fatalf("expected all warehouses row, got warehouse %q", list.Rows[0].WarehouseName)
	}
}

func TestPriceListFiltersByWarehouseQuantity(t *testing.T) {
	fx := testutil.NewFixture(t)
	var warehouse models.Warehouse
	if err := fx.DB.Where("code = ?", "MAIN-WH").First(&warehouse).Error; err != nil {
		t.Fatalf("load warehouse: %v", err)
	}
	if err := fx.DB.Create(&models.ItemWarehouseBalance{ItemID: fx.Item.ID, WarehouseID: warehouse.ID, Quantity: fx.Item.Quantity}).Error; err != nil {
		t.Fatalf("create warehouse balance: %v", err)
	}
	list := NewItemService(fx.DB).PriceList("DEMO-001", "all", warehouse.ID)
	if len(list.Rows) != 1 {
		t.Fatalf("expected one item, got %d", len(list.Rows))
	}
	if list.Rows[0].Quantity != fx.Item.Quantity {
		t.Fatalf("expected warehouse quantity %.3f, got %.3f", fx.Item.Quantity, list.Rows[0].Quantity)
	}
	if list.Rows[0].WarehouseName != warehouse.Name {
		t.Fatalf("expected warehouse %q, got %q", warehouse.Name, list.Rows[0].WarehouseName)
	}
}

func TestSaveItemWithExistingCategory(t *testing.T) {
	fx := testutil.NewFixture(t)
	service := NewItemService(fx.DB)
	tenantID := fx.Tenant.ID

	var category models.ItemCategory
	if err := fx.DB.Where("tenant_id = ? AND name = ?", tenantID, "General").First(&category).Error; err != nil {
		t.Fatalf("load category: %v", err)
	}

	existingItem := models.Item{TenantID: &tenantID, Name: "Category Existing Item", Code: "CAT-EXIST-001", Barcode: "CAT-EXIST-BAR", PurchasePrice: 5, SalePrice: 7, CategoryID: &category.ID}
	if err := service.SaveWithCategory(&existingItem, ""); err != nil {
		t.Fatalf("create item with existing category: %v", err)
	}
	if existingItem.CategoryID == nil || *existingItem.CategoryID != category.ID {
		t.Fatal("expected existing category to be linked")
	}
}

func TestSaveItemRequiresCategoryChoice(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	item := models.Item{TenantID: &tenantID, Name: "No Category Item", Code: "NO-CAT-001", Barcode: "NO-CAT-BAR", PurchasePrice: 5, SalePrice: 7}
	if err := NewItemService(fx.DB).SaveWithCategory(&item, ""); err == nil {
		t.Fatal("expected category validation to fail")
	}
}

func TestCreateItemZeroQuantityWithoutWarehouseSucceeds(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	item := models.Item{TenantID: &tenantID, Name: "Zero Opening Item", Code: "OPEN-ZERO-001", Barcode: "OPEN-ZERO-BAR", PurchasePrice: 5, SalePrice: 7, CategoryID: fx.Item.CategoryID}
	if err := NewItemService(fx.DB).CreateWithOpeningBalance(&item, 0); err != nil {
		t.Fatalf("create zero quantity item: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected item to be created")
	}
}

func TestCreateItemPositiveQuantityWithoutWarehouseFails(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	item := models.Item{TenantID: &tenantID, Name: "Missing Warehouse Item", Code: "OPEN-NOWH-001", Barcode: "OPEN-NOWH-BAR", PurchasePrice: 5, SalePrice: 7, Quantity: 3, CategoryID: fx.Item.CategoryID}
	err := NewItemService(fx.DB).CreateWithOpeningBalance(&item, 0)
	if err == nil {
		t.Fatal("expected missing warehouse to fail")
	}
	if err.Error() != "يجب اختيار المخزن عند إدخال كمية افتتاحية" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateItemPositiveQuantityWithWarehouseCreatesBalanceAndMovement(t *testing.T) {
	fx := testutil.NewFixture(t)
	tenantID := fx.Tenant.ID
	var warehouse models.Warehouse
	if err := fx.DB.Where("code = ?", "MAIN-WH").First(&warehouse).Error; err != nil {
		t.Fatalf("load warehouse: %v", err)
	}
	item := models.Item{TenantID: &tenantID, Name: "Opening Balance Item", Code: "OPEN-BAL-001", Barcode: "OPEN-BAL-BAR", PurchasePrice: 5, SalePrice: 7, Quantity: 4, CategoryID: fx.Item.CategoryID}
	if err := NewItemService(fx.DB).CreateWithOpeningBalance(&item, warehouse.ID); err != nil {
		t.Fatalf("create opening item: %v", err)
	}

	var balance models.ItemWarehouseBalance
	if err := fx.DB.Where("item_id = ? AND warehouse_id = ?", item.ID, warehouse.ID).First(&balance).Error; err != nil {
		t.Fatalf("load warehouse balance: %v", err)
	}
	if balance.Quantity != 4 {
		t.Fatalf("expected balance 4, got %.3f", balance.Quantity)
	}

	var movement models.StockMovement
	if err := fx.DB.Where("item_id = ? AND type = ?", item.ID, models.StockOpen).First(&movement).Error; err != nil {
		t.Fatalf("load opening movement: %v", err)
	}
	if movement.Quantity != 4 || movement.Reference != "رصيد افتتاحي" {
		t.Fatalf("unexpected movement: %+v", movement)
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
