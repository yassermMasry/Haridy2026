package services

import (
	"errors"
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleLineInput struct {
	ItemID    uint
	Quantity  float64
	UnitPrice float64
}

type SaleInput struct {
	UserID      uint
	CustomerID  uint
	PaymentType string
	Discount    float64
	Tax         float64
	PaidCash    float64
	Lines       []SaleLineInput
}

type SalesService struct{ db *gorm.DB }

func NewSalesService(db *gorm.DB) *SalesService { return &SalesService{db: db} }

func (s *SalesService) List() []models.SalesInvoice {
	var invoices []models.SalesInvoice
	s.db.Preload("User").Preload("Customer").Order("created_at desc").Limit(50).Find(&invoices)
	return invoices
}

func (s *SalesService) Find(id uint) (*models.SalesInvoice, error) {
	var invoice models.SalesInvoice
	if err := s.db.Preload("Items.Item").Preload("User").Preload("Customer").First(&invoice, id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *SalesService) Create(input SaleInput) (*models.SalesInvoice, error) {
	if len(input.Lines) == 0 {
		return nil, errors.New("add at least one item")
	}
	var invoice models.SalesInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		invoice = models.SalesInvoice{
			Number:      fmt.Sprintf("INV-%s", time.Now().Format("20060102150405")),
			UserID:      input.UserID,
			PaymentType: input.PaymentType,
			Discount:    input.Discount,
			Tax:         input.Tax,
			PaidCash:    input.PaidCash,
		}
		if invoice.PaymentType == "" {
			invoice.PaymentType = "cash"
		}
		if input.CustomerID > 0 {
			invoice.CustomerID = &input.CustomerID
		}

		var subtotal float64
		for _, line := range input.Lines {
			if line.Quantity <= 0 {
				return errors.New("invalid invoice quantity")
			}
			var item models.Item
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, line.ItemID).Error; err != nil {
				return err
			}
			if err := ValidateInventoryDelta(item.Quantity, -line.Quantity, item.Name); err != nil {
				return err
			}
			price := line.UnitPrice
			if price <= 0 {
				price = item.SalePrice
			}
			total := price * line.Quantity
			subtotal += total
			invoice.Items = append(invoice.Items, models.SalesInvoiceItem{ItemID: item.ID, Quantity: line.Quantity, UnitPrice: price, Total: total})
			if err := tx.Model(&item).Update("quantity", gorm.Expr("quantity - ?", line.Quantity)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.StockMovement{ItemID: item.ID, Type: models.StockSale, Quantity: line.Quantity, Reference: invoice.Number, PerformedBy: &input.UserID}).Error; err != nil {
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
			if input.CustomerID == 0 {
				return errors.New("select a customer for credit sales")
			}
			var customer models.Customer
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&customer, input.CustomerID).Error; err != nil {
				return err
			}
			if customer.CreditLimit > 0 && customer.Balance+invoice.Total > customer.CreditLimit {
				return errors.New("customer credit limit exceeded")
			}
			if err := tx.Model(&customer).Update("balance", gorm.Expr("balance + ?", invoice.Total)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.CustomerTransaction{CustomerID: customer.ID, Type: models.CustomerSale, Debit: invoice.Total, Reference: invoice.Number, Description: "Credit sale", UserID: &input.UserID}).Error; err != nil {
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
			if err := ValidateNonNegativeBalance(treasury.Balance, invoice.PaidCash, "treasury balance"); err != nil {
				return err
			}
			if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance + ?", invoice.PaidCash)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.TreasuryTransaction{TreasuryID: treasury.ID, Type: models.TreasurySale, Amount: invoice.PaidCash, Reference: invoice.Number, Description: "Cash sale", UserID: &input.UserID}).Error; err != nil {
				return err
			}
		}
		debitAccount := "1000"
		if invoice.PaymentType == "credit" {
			debitAccount = "1200"
		}
		if err := CreateJournal(tx, invoice.Number, "Sales invoice", []JournalLineInput{{AccountCode: debitAccount, Debit: invoice.Total}, {AccountCode: "4000", Credit: invoice.Total}}); err != nil {
			return err
		}
		return Audit(tx, input.UserID, "CREATE", "sales_invoices", invoice.ID, invoice.Number)
	})
	return &invoice, err
}
