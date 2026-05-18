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
