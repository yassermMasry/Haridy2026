package services

import (
	"errors"
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type ReturnService struct{ db *gorm.DB }

func NewReturnService(db *gorm.DB) *ReturnService { return &ReturnService{db: db} }

func (s *ReturnService) SalesReturn(invoiceID uint, reason string, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var invoice models.SalesInvoice
		if err := tx.Preload("Items").First(&invoice, invoiceID).Error; err != nil {
			return err
		}
		if len(invoice.Items) == 0 {
			return errors.New("invoice has no items")
		}
		ret := models.SalesReturn{Number: fmt.Sprintf("SR-%s", time.Now().Format("20060102150405")), InvoiceID: invoice.ID, BranchID: invoice.BranchID, WarehouseID: invoice.WarehouseID, Total: invoice.Total, Reason: reason, UserID: userID}
		for _, line := range invoice.Items {
			ret.Items = append(ret.Items, models.SalesReturnItem{ItemID: line.ItemID, Quantity: line.Quantity, UnitPrice: line.UnitPrice, Total: line.Total})
			if err := tx.Model(&models.Item{}).Where("id = ?", line.ItemID).Update("quantity", gorm.Expr("quantity + ?", line.Quantity)).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&ret).Error; err != nil {
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
		}
		if err := CreateJournal(tx, ret.Number, "Sales return", []JournalLineInput{{AccountCode: "4000", Debit: invoice.Total}, {AccountCode: "1000", Credit: invoice.Total}}); err != nil {
			return err
		}
		return Audit(tx, userID, "CREATE", "sales_returns", ret.ID, ret.Number)
	})
}

func (s *ReturnService) PurchaseReturn(invoiceID uint, reason string, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var invoice models.PurchaseInvoice
		if err := tx.Preload("Items").First(&invoice, invoiceID).Error; err != nil {
			return err
		}
		ret := models.PurchaseReturn{Number: fmt.Sprintf("PR-%s", time.Now().Format("20060102150405")), InvoiceID: invoice.ID, BranchID: invoice.BranchID, WarehouseID: invoice.WarehouseID, Total: invoice.Total, Reason: reason, UserID: userID}
		for _, line := range invoice.Items {
			ret.Items = append(ret.Items, models.PurchaseReturnItem{ItemID: line.ItemID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: line.Total})
			if err := tx.Model(&models.Item{}).Where("id = ?", line.ItemID).Update("quantity", gorm.Expr("quantity - ?", line.Quantity)).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&ret).Error; err != nil {
			return err
		}
		if invoice.PaidCash > 0 {
			var treasury models.Treasury
			if err := tx.First(&treasury).Error; err != nil {
				return err
			}
			if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance + ?", invoice.PaidCash)).Error; err != nil {
				return err
			}
		}
		if err := CreateJournal(tx, ret.Number, "Purchase return", []JournalLineInput{{AccountCode: "1000", Debit: invoice.Total}, {AccountCode: "5000", Credit: invoice.Total}}); err != nil {
			return err
		}
		return Audit(tx, userID, "CREATE", "purchase_returns", ret.ID, ret.Number)
	})
}
