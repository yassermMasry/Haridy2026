package services

import (
	"errors"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type SupplierService struct{ db *gorm.DB }

func NewSupplierService(db *gorm.DB) *SupplierService { return &SupplierService{db: db} }

func (s *SupplierService) List() []models.Supplier {
	var suppliers []models.Supplier
	s.db.Order("created_at desc").Find(&suppliers)
	return suppliers
}

func (s *SupplierService) Find(id uint) (*models.Supplier, []models.SupplierTransaction, error) {
	var supplier models.Supplier
	if err := s.db.First(&supplier, id).Error; err != nil {
		return nil, nil, err
	}
	var txs []models.SupplierTransaction
	s.db.Where("supplier_id = ?", id).Order("created_at desc").Find(&txs)
	return &supplier, txs, nil
}

func (s *SupplierService) Save(supplier *models.Supplier, userID uint) error {
	if supplier.Name == "" {
		return errors.New("اسم المورد مطلوب")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if supplier.ID == 0 {
			if err := tx.Create(supplier).Error; err != nil {
				return err
			}
			return Audit(tx, userID, "CREATE", "suppliers", supplier.ID, supplier.Name)
		}
		if err := tx.Model(&models.Supplier{}).Where("id = ?", supplier.ID).Updates(supplier).Error; err != nil {
			return err
		}
		return Audit(tx, userID, "UPDATE", "suppliers", supplier.ID, supplier.Name)
	})
}

func (s *SupplierService) Pay(supplierID uint, amount float64, userID uint) error {
	if amount <= 0 {
		return errors.New("المبلغ يجب أن يكون أكبر من صفر")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var supplier models.Supplier
		if err := tx.First(&supplier, supplierID).Error; err != nil {
			return err
		}
		if err := tx.Model(&supplier).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.SupplierTransaction{SupplierID: supplier.ID, Type: models.SupplierPayment, Debit: amount, Description: "دفعة مورد", UserID: &userID}).Error; err != nil {
			return err
		}
		var treasury models.Treasury
		if err := tx.First(&treasury).Error; err != nil {
			return err
		}
		if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.TreasuryTransaction{TreasuryID: treasury.ID, Type: models.TreasurySupplierPay, Amount: amount, Description: "سداد مورد", UserID: &userID}).Error; err != nil {
			return err
		}
		if err := CreateJournal(tx, "SUPPLIER-PAYMENT", "سداد مورد", []JournalLineInput{{AccountCode: "2000", Debit: amount}, {AccountCode: "1000", Credit: amount}}); err != nil {
			return err
		}
		return Audit(tx, userID, "PAY", "suppliers", supplier.ID, supplier.Name)
	})
}
