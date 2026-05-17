package services

import (
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type JournalLineInput struct {
	AccountCode string
	Debit       float64
	Credit      float64
}

func CreateJournal(tx *gorm.DB, reference, description string, lines []JournalLineInput) error {
	entry := models.JournalEntry{
		Number:      fmt.Sprintf("JE-%s", time.Now().Format("20060102150405.000000")),
		Reference:   reference,
		Description: description,
	}
	for _, line := range lines {
		var account models.ChartOfAccount
		if err := tx.Where("code = ?", line.AccountCode).First(&account).Error; err != nil {
			return err
		}
		entry.Lines = append(entry.Lines, models.JournalEntryLine{
			AccountID: account.ID,
			Debit:     line.Debit,
			Credit:    line.Credit,
		})
	}
	return tx.Create(&entry).Error
}

func Audit(tx *gorm.DB, userID uint, action, entity string, entityID uint, details string) error {
	var uid *uint
	if userID > 0 {
		uid = &userID
	}
	return tx.Create(&models.AuditLog{UserID: uid, Action: action, Entity: entity, EntityID: entityID, Details: details}).Error
}

type AccountingService struct{ db *gorm.DB }

func NewAccountingService(db *gorm.DB) *AccountingService { return &AccountingService{db: db} }

func (s *AccountingService) Accounts() []models.ChartOfAccount {
	var accounts []models.ChartOfAccount
	s.db.Order("code asc").Find(&accounts)
	return accounts
}

func (s *AccountingService) Journal() []models.JournalEntry {
	var entries []models.JournalEntry
	s.db.Preload("Lines.Account").Order("created_at desc").Limit(100).Find(&entries)
	return entries
}

func (s *AccountingService) AuditLogs() []models.AuditLog {
	var logs []models.AuditLog
	s.db.Order("created_at desc").Limit(100).Find(&logs)
	return logs
}
