package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"time"

	"haridy2026/internal/models"

	"gorm.io/gorm"
)

type CommercialService struct{ db *gorm.DB }

func NewCommercialService(db *gorm.DB) *CommercialService { return &CommercialService{db: db} }

func (s *CommercialService) OnboardCompany(name, slug, ownerUsername, passwordHash string) (*models.Tenant, error) {
	var plan models.Plan
	if err := s.db.Where("code = ?", "starter").First(&plan).Error; err != nil {
		return nil, err
	}
	var tenant models.Tenant
	err := s.db.Transaction(func(tx *gorm.DB) error {
		tenant = models.Tenant{Name: name, Slug: slug, Subdomain: slug, Status: "trial"}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.CompanySetting{TenantID: tenant.ID, LegalName: name, Currency: "EGP", TimeZone: "Africa/Cairo"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.Subscription{TenantID: tenant.ID, PlanID: plan.ID, Status: "trial", StartsAt: time.Now()}).Error; err != nil {
			return err
		}
		user := models.User{TenantID: &tenant.ID, Username: ownerUsername, PasswordHash: passwordHash, Role: models.RoleAdmin}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.TenantUser{TenantID: tenant.ID, UserID: user.ID, IsOwner: true}).Error; err != nil {
			return err
		}
		return nil
	})
	return &tenant, err
}

func (s *CommercialService) TrialBalance(tenantID uint) []map[string]any {
	var rows []struct {
		Code   string
		Name   string
		Debit  float64
		Credit float64
	}
	s.db.Table("chart_of_accounts").
		Select("chart_of_accounts.code, chart_of_accounts.name, COALESCE(SUM(journal_entry_lines.debit),0) debit, COALESCE(SUM(journal_entry_lines.credit),0) credit").
		Joins("LEFT JOIN journal_entry_lines ON journal_entry_lines.account_id = chart_of_accounts.id").
		Where("chart_of_accounts.tenant_id = ? OR chart_of_accounts.tenant_id IS NULL", tenantID).
		Group("chart_of_accounts.code, chart_of_accounts.name").
		Order("chart_of_accounts.code asc").
		Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{"code": row.Code, "name": row.Name, "debit": row.Debit, "credit": row.Credit, "balance": row.Debit - row.Credit})
	}
	return out
}

func (s *CommercialService) FinancialSummary(tenantID uint) map[string]float64 {
	var sales, purchases, cash, receivables, payables float64
	s.db.Model(&models.SalesInvoice{}).Where("tenant_id = ? OR tenant_id IS NULL", tenantID).Select("COALESCE(SUM(total),0)").Scan(&sales)
	s.db.Model(&models.PurchaseInvoice{}).Where("tenant_id = ? OR tenant_id IS NULL", tenantID).Select("COALESCE(SUM(total),0)").Scan(&purchases)
	s.db.Model(&models.Treasury{}).Where("tenant_id = ? OR tenant_id IS NULL", tenantID).Select("COALESCE(SUM(balance),0)").Scan(&cash)
	s.db.Model(&models.Customer{}).Where("tenant_id = ? OR tenant_id IS NULL", tenantID).Select("COALESCE(SUM(balance),0)").Scan(&receivables)
	s.db.Model(&models.Supplier{}).Where("tenant_id = ? OR tenant_id IS NULL", tenantID).Select("COALESCE(SUM(balance),0)").Scan(&payables)
	return map[string]float64{"sales": sales, "purchases": purchases, "gross_profit": sales - purchases, "cash": cash, "receivables": receivables, "payables": payables, "assets": cash + receivables, "liabilities": payables}
}

func (s *CommercialService) GenerateEInvoice(tenantID, invoiceID uint) (*models.EInvoice, error) {
	var inv models.SalesInvoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, err
	}
	payload := fmt.Sprintf("tenant=%d;invoice=%s;total=%.2f;vat=%.2f", tenantID, inv.Number, inv.Total, inv.Tax)
	sum := sha256.Sum256([]byte(payload))
	doc := struct {
		XMLName xml.Name `xml:"Invoice"`
		Number  string   `xml:"Number"`
		Total   float64  `xml:"Total"`
		VAT     float64  `xml:"VAT"`
	}{Number: inv.Number, Total: inv.Total, VAT: inv.Tax}
	raw, _ := xml.MarshalIndent(doc, "", "  ")
	einv := models.EInvoice{TenantID: tenantID, InvoiceID: invoiceID, UUID: hex.EncodeToString(sum[:16]), QRPayload: payload, XMLBody: string(raw), VATAmount: inv.Tax, Status: "generated"}
	return &einv, s.db.Create(&einv).Error
}

func (s *CommercialService) CloseFiscalYear(tenantID, fiscalYearID, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		entry := models.JournalEntry{TenantID: &tenantID, Number: fmt.Sprintf("CLOSE-%d-%d", fiscalYearID, time.Now().Unix()), Reference: "YEAR-CLOSE", Description: "Fiscal year closing"}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.ClosingEntry{TenantID: tenantID, FiscalYearID: fiscalYearID, JournalEntryID: entry.ID}).Error; err != nil {
			return err
		}
		return tx.Model(&models.FiscalYear{}).Where("id = ? AND tenant_id = ?", fiscalYearID, tenantID).Update("is_closed", true).Error
	})
}
