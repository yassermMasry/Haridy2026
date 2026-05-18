package services

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

const moneyEpsilon = 0.005

func ValidateJournalLines(lines []JournalLineInput) error {
	if len(lines) < 2 {
		return errors.New("journal entry must contain at least two lines")
	}
	var debit, credit float64
	for _, line := range lines {
		if line.AccountCode == "" {
			return errors.New("journal account code is required")
		}
		if line.Debit < 0 || line.Credit < 0 {
			return errors.New("journal line cannot contain negative amounts")
		}
		if line.Debit > 0 && line.Credit > 0 {
			return errors.New("journal line cannot debit and credit the same account")
		}
		if line.Debit == 0 && line.Credit == 0 {
			return errors.New("journal line amount is required")
		}
		debit += line.Debit
		credit += line.Credit
	}
	if math.Abs(debit-credit) > moneyEpsilon {
		return fmt.Errorf("journal entry is not balanced: debit %.2f credit %.2f", debit, credit)
	}
	return nil
}

func ValidatePostingNotDuplicate(tx *gorm.DB, reference string) error {
	if reference == "" {
		return errors.New("posting reference is required")
	}
	var count int64
	if err := tx.Model(&models.JournalEntry{}).Where("reference = ?", reference).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("duplicate posting for reference %s", reference)
	}
	return nil
}

func ValidateNonNegativeBalance(current, delta float64, label string) error {
	if current+delta < -moneyEpsilon {
		return fmt.Errorf("%s cannot become negative", label)
	}
	return nil
}

func ValidateInventoryDelta(current, delta float64, itemName string) error {
	if current+delta < -0.0005 {
		return fmt.Errorf("inventory corruption prevented for %s", itemName)
	}
	return nil
}

func ReconcileAccounting(db *gorm.DB) (string, error) {
	var findings []string
	var entries []models.JournalEntry
	if err := db.Preload("Lines").Find(&entries).Error; err != nil {
		return "", err
	}
	for _, entry := range entries {
		var debit, credit float64
		for _, line := range entry.Lines {
			if line.Debit < 0 || line.Credit < 0 || (line.Debit > 0 && line.Credit > 0) {
				findings = append(findings, fmt.Sprintf("entry %s has invalid line amounts", entry.Number))
			}
			debit += line.Debit
			credit += line.Credit
		}
		if math.Abs(debit-credit) > moneyEpsilon {
			findings = append(findings, fmt.Sprintf("entry %s unbalanced debit %.2f credit %.2f", entry.Number, debit, credit))
		}
	}
	return strings.Join(findings, "\n"), nil
}

func ReconcileInventory(db *gorm.DB) (string, error) {
	var findings []string
	var items []models.Item
	if err := db.Find(&items).Error; err != nil {
		return "", err
	}
	for _, item := range items {
		if item.Quantity < -0.0005 {
			findings = append(findings, fmt.Sprintf("item %s has negative quantity %.3f", item.Code, item.Quantity))
		}
		var movedIn, movedOut float64
		if err := db.Model(&models.StockMovement{}).Where("item_id = ? AND type IN ?", item.ID, []models.StockMovementType{models.StockIn, models.StockBuy}).Select("COALESCE(SUM(quantity),0)").Scan(&movedIn).Error; err != nil {
			return "", err
		}
		if err := db.Model(&models.StockMovement{}).Where("item_id = ? AND type IN ?", item.ID, []models.StockMovementType{models.StockOut, models.StockSale}).Select("COALESCE(SUM(quantity),0)").Scan(&movedOut).Error; err != nil {
			return "", err
		}
		if movedOut-movedIn > item.Quantity+0.0005 {
			findings = append(findings, fmt.Sprintf("item %s movements exceed available stock", item.Code))
		}
	}
	return strings.Join(findings, "\n"), nil
}

func RunReconciliation(db *gorm.DB, scope string, tenantID *uint) error {
	started := time.Now()
	run := models.ReconciliationRun{TenantID: tenantID, Scope: scope, Status: "running", StartedAt: started}
	if err := db.Create(&run).Error; err != nil {
		return err
	}
	var findings string
	var err error
	switch scope {
	case "accounting":
		findings, err = ReconcileAccounting(db)
	case "inventory":
		findings, err = ReconcileInventory(db)
	default:
		accounting, accErr := ReconcileAccounting(db)
		inventory, invErr := ReconcileInventory(db)
		findings = strings.TrimSpace(accounting + "\n" + inventory)
		if accErr != nil {
			err = accErr
		} else {
			err = invErr
		}
	}
	completed := time.Now()
	status := "passed"
	if err != nil {
		status = "failed"
		findings = err.Error()
	} else if strings.TrimSpace(findings) != "" {
		status = "findings"
	}
	return db.Model(&run).Updates(map[string]any{"status": status, "findings": findings, "completed_at": &completed}).Error
}
