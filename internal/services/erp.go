package services

import (
	"errors"
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ERPService struct{ db *gorm.DB }

func NewERPService(db *gorm.DB) *ERPService { return &ERPService{db: db} }

func (s *ERPService) Branches() []models.Branch {
	var branches []models.Branch
	s.db.Order("name asc").Find(&branches)
	return branches
}

func (s *ERPService) Warehouses() []models.Warehouse {
	var warehouses []models.Warehouse
	s.db.Preload("Branch").Order("name asc").Find(&warehouses)
	return warehouses
}

func (s *ERPService) Transfer(itemID, fromWarehouseID, toWarehouseID uint, qty float64, userID uint) error {
	if itemID == 0 || fromWarehouseID == 0 || toWarehouseID == 0 || qty <= 0 || fromWarehouseID == toWarehouseID {
		return errors.New("invalid warehouse transfer")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var from models.ItemWarehouseBalance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("item_id = ? AND warehouse_id = ?", itemID, fromWarehouseID).First(&from).Error; err != nil {
			return err
		}
		if from.Quantity < qty {
			return errors.New("insufficient warehouse balance")
		}
		if err := tx.Model(&from).Update("quantity", gorm.Expr("quantity - ?", qty)).Error; err != nil {
			return err
		}
		var to models.ItemWarehouseBalance
		if err := tx.Where("item_id = ? AND warehouse_id = ?", itemID, toWarehouseID).FirstOrCreate(&to, models.ItemWarehouseBalance{ItemID: itemID, WarehouseID: toWarehouseID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&to).Update("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
			return err
		}
		transfer := models.WarehouseTransfer{Number: fmt.Sprintf("TR-%s", time.Now().Format("20060102150405")), ItemID: itemID, FromWarehouseID: fromWarehouseID, ToWarehouseID: toWarehouseID, Quantity: qty, UserID: userID}
		if err := tx.Create(&transfer).Error; err != nil {
			return err
		}
		return Audit(tx, userID, "TRANSFER", "warehouse_transfers", transfer.ID, transfer.Number)
	})
}

func (s *ERPService) Notifications(branchID uint) []models.Notification {
	var notes []models.Notification
	q := s.db.Order("created_at desc").Limit(20)
	if branchID > 0 {
		q = q.Where("branch_id IS NULL OR branch_id = ?", branchID)
	}
	q.Find(&notes)
	return notes
}

func (s *ERPService) GenerateNotifications() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var lowStock []models.Item
		tx.Where("quantity <= minimum_stock").Limit(25).Find(&lowStock)
		for _, item := range lowStock {
			title := "Low stock: " + item.Name
			var count int64
			tx.Model(&models.Notification{}).Where("type = ? AND title = ? AND is_read = ?", "stock", title, false).Count(&count)
			if count == 0 {
				if err := tx.Create(&models.Notification{Type: "stock", Title: title, Message: fmt.Sprintf("Current quantity %.3f", item.Quantity)}).Error; err != nil {
					return err
				}
			}
		}
		var customers []models.Customer
		tx.Where("balance > 0").Limit(25).Find(&customers)
		for _, customer := range customers {
			title := "Customer debt: " + customer.Name
			var count int64
			tx.Model(&models.Notification{}).Where("type = ? AND title = ? AND is_read = ?", "debt", title, false).Count(&count)
			if count == 0 {
				if err := tx.Create(&models.Notification{BranchID: customer.BranchID, Type: "debt", Title: title, Message: fmt.Sprintf("Balance %.2f", customer.Balance)}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *ERPService) ReceiptVoucher(customerID uint, amount float64, description string, userID uint) (*models.ReceiptVoucher, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}
	var voucher models.ReceiptVoucher
	err := s.db.Transaction(func(tx *gorm.DB) error {
		voucher = models.ReceiptVoucher{Number: fmt.Sprintf("RV-%s", time.Now().Format("20060102150405")), CustomerID: &customerID, Amount: amount, Description: description, UserID: userID}
		if err := tx.Create(&voucher).Error; err != nil {
			return err
		}
		cs := NewCustomerService(tx)
		if err := cs.Receive(customerID, amount, userID); err != nil {
			return err
		}
		return Audit(tx, userID, "CREATE", "receipt_vouchers", voucher.ID, voucher.Number)
	})
	return &voucher, err
}

func (s *ERPService) PaymentVoucher(supplierID uint, amount float64, description string, userID uint) (*models.PaymentVoucher, error) {
	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}
	var voucher models.PaymentVoucher
	err := s.db.Transaction(func(tx *gorm.DB) error {
		voucher = models.PaymentVoucher{Number: fmt.Sprintf("PV-%s", time.Now().Format("20060102150405")), SupplierID: &supplierID, Amount: amount, Description: description, UserID: userID}
		if err := tx.Create(&voucher).Error; err != nil {
			return err
		}
		ss := NewSupplierService(tx)
		if err := ss.Pay(supplierID, amount, userID); err != nil {
			return err
		}
		return Audit(tx, userID, "CREATE", "payment_vouchers", voucher.ID, voucher.Number)
	})
	return &voucher, err
}
