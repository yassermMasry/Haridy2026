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
	TenantID    *uint
	UserID      uint
	CustomerID  uint
	WarehouseID uint
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
	if input.PaidCash < 0 {
		return nil, errors.New("المدفوع لا يكون سالبًا")
	}
	if err := NewUsageLimitService(s.db).CheckOperationLimit(input.TenantID); err != nil {
		return nil, err
	}
	var invoice models.SalesInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		invoice = models.SalesInvoice{
			TenantID:    input.TenantID,
			Number:      fmt.Sprintf("INV-%s", time.Now().Format("20060102150405")),
			UserID:      input.UserID,
			PaymentType: input.PaymentType,
			Discount:    input.Discount,
			Tax:         input.Tax,
			PaidCash:    input.PaidCash,
		}
		if input.WarehouseID > 0 {
			invoice.WarehouseID = &input.WarehouseID
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
			warehouseID, available, err := s.saleWarehouseBalance(tx, item.ID, input.WarehouseID)
			if err != nil {
				return err
			}
			if available == 0 {
				available = item.Quantity
			}
			if err := ValidateInventoryDelta(available, -line.Quantity, item.Name); err != nil {
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
			if warehouseID > 0 {
				if err := tx.Model(&models.ItemWarehouseBalance{}).
					Where("item_id = ? AND warehouse_id = ?", item.ID, warehouseID).
					Update("quantity", gorm.Expr("quantity - ?", line.Quantity)).Error; err != nil {
					return err
				}
			}
			movement := models.StockMovement{TenantID: input.TenantID, ItemID: item.ID, Type: models.StockSale, Quantity: line.Quantity, Reference: invoice.Number, PerformedBy: &input.UserID}
			if warehouseID > 0 {
				movement.WarehouseID = &warehouseID
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		invoice.Subtotal = subtotal
		invoice.Total = subtotal - invoice.Discount + invoice.Tax
		if invoice.PaymentType == "cash" && invoice.PaidCash <= 0 {
			invoice.PaidCash = invoice.Total
		}
		remaining := invoice.Total - invoice.PaidCash
		if remaining != 0 && input.CustomerID == 0 {
			return errors.New("لا يسمح بآجل أو زيادة بدون عميل")
		}
		if invoice.PaymentType == "credit" && input.CustomerID == 0 {
			return errors.New("select a customer for credit sales")
		}
		if remaining > 0 {
			if input.CustomerID == 0 {
				return errors.New("select a customer for credit sales")
			}
			var customer models.Customer
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&customer, input.CustomerID).Error; err != nil {
				return err
			}
			if customer.CreditLimit > 0 && customer.Balance+remaining > customer.CreditLimit {
				return errors.New("customer credit limit exceeded")
			}
			if err := tx.Model(&customer).Update("balance", gorm.Expr("balance + ?", remaining)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.CustomerTransaction{CustomerID: customer.ID, Type: models.CustomerSale, Debit: remaining, Reference: invoice.Number, Description: "Credit sale", UserID: &input.UserID}).Error; err != nil {
				return err
			}
		}
		if remaining < 0 {
			credit := -remaining
			var customer models.Customer
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&customer, input.CustomerID).Error; err != nil {
				return err
			}
			if err := tx.Model(&customer).Update("balance", gorm.Expr("balance - ?", credit)).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.CustomerTransaction{CustomerID: customer.ID, Type: models.CustomerPayment, Credit: credit, Reference: invoice.Number, Description: "Overpaid sale credit", UserID: &input.UserID}).Error; err != nil {
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
		lines := []JournalLineInput{{AccountCode: "4000", Credit: invoice.Total}}
		if invoice.PaidCash > 0 {
			lines = append(lines, JournalLineInput{AccountCode: "1000", Debit: invoice.PaidCash})
		}
		if remaining > 0 {
			lines = append(lines, JournalLineInput{AccountCode: "1200", Debit: remaining})
		}
		if remaining < 0 {
			lines = append(lines, JournalLineInput{AccountCode: "1200", Credit: -remaining})
		}
		if err := CreateJournal(tx, invoice.Number, "Sales invoice", lines); err != nil {
			return err
		}
		return Audit(tx, input.UserID, "CREATE", "sales_invoices", invoice.ID, invoice.Number)
	})
	return &invoice, err
}

func (s *SalesService) saleWarehouseBalance(tx *gorm.DB, itemID, requestedWarehouseID uint) (uint, float64, error) {
	if requestedWarehouseID > 0 {
		var balance models.ItemWarehouseBalance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("item_id = ? AND warehouse_id = ?", itemID, requestedWarehouseID).
			First(&balance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return requestedWarehouseID, 0, nil
			}
			return 0, 0, err
		}
		return requestedWarehouseID, balance.Quantity, nil
	}
	var balances []models.ItemWarehouseBalance
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("item_id = ? AND quantity > 0", itemID).
		Find(&balances).Error; err != nil {
		return 0, 0, err
	}
	if len(balances) > 1 {
		return 0, 0, errors.New("يجب تحديد المخزن عند وجود الصنف في أكثر من مخزن")
	}
	if len(balances) == 1 {
		return balances[0].WarehouseID, balances[0].Quantity, nil
	}
	return 0, 0, nil
}
