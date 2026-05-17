package services

import (
	"errors"
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseLineInput struct {
	ItemID   uint
	Quantity float64
	UnitCost float64
}

type PurchaseInput struct {
	UserID      uint
	SupplierID  uint
	PaymentType string
	Discount    float64
	Tax         float64
	PaidCash    float64
	Lines       []PurchaseLineInput
}

type PurchaseService struct{ db *gorm.DB }

func NewPurchaseService(db *gorm.DB) *PurchaseService { return &PurchaseService{db: db} }

func (s *PurchaseService) List() []models.PurchaseInvoice {
	var invoices []models.PurchaseInvoice
	s.db.Preload("Supplier").Order("created_at desc").Limit(50).Find(&invoices)
	return invoices
}

func (s *PurchaseService) Find(id uint) (*models.PurchaseInvoice, error) {
	var invoice models.PurchaseInvoice
	if err := s.db.Preload("Items.Item").Preload("Supplier").First(&invoice, id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *PurchaseService) Create(input PurchaseInput) (*models.PurchaseInvoice, error) {
	if input.SupplierID == 0 {
		return nil, errors.New("select supplier")
	}
	if len(input.Lines) == 0 {
		return nil, errors.New("add at least one item")
	}
	var invoice models.PurchaseInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		invoice = models.PurchaseInvoice{Number: fmt.Sprintf("PINV-%s", time.Now().Format("20060102150405")), SupplierID: input.SupplierID, UserID: input.UserID, PaymentType: input.PaymentType, Discount: input.Discount, Tax: input.Tax, PaidCash: input.PaidCash}
		if invoice.PaymentType == "" {
			invoice.PaymentType = "cash"
		}
		var subtotal float64
		for _, line := range input.Lines {
			if line.Quantity <= 0 || line.UnitCost < 0 {
				return errors.New("invalid purchase line")
			}
			var item models.Item
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, line.ItemID).Error; err != nil {
				return err
			}
			total := line.Quantity * line.UnitCost
			subtotal += total
			invoice.Items = append(invoice.Items, models.PurchaseInvoiceItem{ItemID: item.ID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: total})
			if err := tx.Model(&item).Updates(map[string]any{"quantity": gorm.Expr("quantity + ?", line.Quantity), "purchase_price": line.UnitCost}).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.StockMovement{ItemID: item.ID, Type: models.StockBuy, Quantity: line.Quantity, Reference: invoice.Number, PerformedBy: &input.UserID}).Error; err != nil {
				return err
			}
		}
		invoice.Subtotal = subtotal
		invoice.Total = subtotal - invoice.Discount + invoice.Tax
		if invoice.PaymentType == "cash" && invoice.PaidCash <= 0 {
			invoice.PaidCash = invoice.Total
		}
		if invoice.PaymentType == "credit" {
			invoice.PaidCash = 0
			var supplier models.Supplier
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&supplier, input.SupplierID).Error; err != nil {
				return err
			}
			if err := tx.Model(&supplier).Update("balance", gorm.Expr("balance + ?", invoice.Total)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.SupplierTransaction{SupplierID: supplier.ID, Type: models.SupplierPurchase, Credit: invoice.Total, Reference: invoice.Number, Description: "Credit purchase", UserID: &input.UserID}).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}
		if invoice.PaidCash > 0 {
			var treasury models.Treasury
			if err := tx.First(&treasury).Error; err != nil {
				return err
			}
			if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance - ?", invoice.PaidCash)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.TreasuryTransaction{TreasuryID: treasury.ID, Type: models.TreasuryPurchase, Amount: invoice.PaidCash, Reference: invoice.Number, Description: "Cash purchase", UserID: &input.UserID}).Error; err != nil {
				return err
			}
		}
		creditAccount := "1000"
		if invoice.PaymentType == "credit" {
			creditAccount = "2000"
		}
		if err := CreateJournal(tx, invoice.Number, "Purchase invoice", []JournalLineInput{{AccountCode: "5000", Debit: invoice.Total}, {AccountCode: creditAccount, Credit: invoice.Total}}); err != nil {
			return err
		}
		return Audit(tx, input.UserID, "CREATE", "purchase_invoices", invoice.ID, invoice.Number)
	})
	return &invoice, err
}
