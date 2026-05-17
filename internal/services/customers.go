package services

import (
	"errors"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type CustomerService struct{ db *gorm.DB }

func NewCustomerService(db *gorm.DB) *CustomerService { return &CustomerService{db: db} }

func (s *CustomerService) List() []models.Customer {
	var customers []models.Customer
	s.db.Order("created_at desc").Find(&customers)
	return customers
}

func (s *CustomerService) Find(id uint) (*models.Customer, []models.CustomerTransaction, error) {
	var customer models.Customer
	if err := s.db.First(&customer, id).Error; err != nil {
		return nil, nil, err
	}
	var txs []models.CustomerTransaction
	s.db.Where("customer_id = ?", id).Order("created_at desc").Find(&txs)
	return &customer, txs, nil
}

func (s *CustomerService) Save(customer *models.Customer, userID uint) error {
	if customer.Name == "" {
		return errors.New("اسم العميل مطلوب")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if customer.ID == 0 {
			if err := tx.Create(customer).Error; err != nil {
				return err
			}
			return Audit(tx, userID, "CREATE", "customers", customer.ID, customer.Name)
		}
		if err := tx.Model(&models.Customer{}).Where("id = ?", customer.ID).Updates(customer).Error; err != nil {
			return err
		}
		return Audit(tx, userID, "UPDATE", "customers", customer.ID, customer.Name)
	})
}

func (s *CustomerService) Receive(customerID uint, amount float64, userID uint) error {
	if amount <= 0 {
		return errors.New("المبلغ يجب أن يكون أكبر من صفر")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.First(&customer, customerID).Error; err != nil {
			return err
		}
		if err := tx.Model(&customer).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.CustomerTransaction{CustomerID: customer.ID, Type: models.CustomerPayment, Credit: amount, Description: "تحصيل دفعة", UserID: &userID}).Error; err != nil {
			return err
		}
		var treasury models.Treasury
		if err := tx.First(&treasury).Error; err != nil {
			return err
		}
		if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.TreasuryTransaction{TreasuryID: treasury.ID, Type: models.TreasuryReceive, Amount: amount, Description: "تحصيل من عميل", UserID: &userID}).Error; err != nil {
			return err
		}
		if err := CreateJournal(tx, "CUSTOMER-PAYMENT", "تحصيل من عميل", []JournalLineInput{{AccountCode: "1000", Debit: amount}, {AccountCode: "1200", Credit: amount}}); err != nil {
			return err
		}
		return Audit(tx, userID, "RECEIVE", "customers", customer.ID, customer.Name)
	})
}
