package services

import (
	"errors"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type TreasuryService struct{ db *gorm.DB }

func NewTreasuryService(db *gorm.DB) *TreasuryService { return &TreasuryService{db: db} }

func (s *TreasuryService) Main() models.Treasury {
	var treasury models.Treasury
	s.db.FirstOrCreate(&treasury, models.Treasury{Name: "الخزينة الرئيسية"})
	return treasury
}

func (s *TreasuryService) Transactions() ([]models.TreasuryTransaction, models.Treasury) {
	treasury := s.Main()
	var txs []models.TreasuryTransaction
	s.db.Preload("User").Where("treasury_id = ?", treasury.ID).Order("created_at desc").Limit(100).Find(&txs)
	return txs, treasury
}

func (s *TreasuryService) Add(tType models.TreasuryTransactionType, amount float64, description string, userID uint) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var treasury models.Treasury
		if err := tx.First(&treasury).Error; err != nil {
			return err
		}
		delta := amount
		if tType == models.TreasuryExpense {
			delta = -amount
		}
		if err := tx.Model(&treasury).Update("balance", gorm.Expr("balance + ?", delta)).Error; err != nil {
			return err
		}
		entry := models.TreasuryTransaction{TreasuryID: treasury.ID, Type: tType, Amount: amount, Description: description, UserID: &userID}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		lines := []JournalLineInput{{AccountCode: "1000", Debit: amount}, {AccountCode: "4000", Credit: amount}}
		if tType == models.TreasuryExpense {
			lines = []JournalLineInput{{AccountCode: "5000", Debit: amount}, {AccountCode: "1000", Credit: amount}}
		}
		if err := CreateJournal(tx, string(tType), description, lines); err != nil {
			return err
		}
		return Audit(tx, userID, "CREATE", "treasury_transactions", entry.ID, description)
	})
}
