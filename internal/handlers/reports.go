package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db         *gorm.DB
	accounting *services.AccountingService
}

func NewReportHandler(db *gorm.DB, accounting *services.AccountingService) *ReportHandler {
	return &ReportHandler{db: db, accounting: accounting}
}

func (h *ReportHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["journal"] = h.accounting.Journal()
	view["accounts"] = h.accounting.Accounts()
	view["auditLogs"] = h.accounting.AuditLogs()
	c.HTML(http.StatusOK, "reports/index.html", view)
}

func (h *ReportHandler) Export(c *gin.Context) {
	name := c.Param("name")
	format := c.DefaultQuery("format", "excel")
	rows := h.rows(name)
	if format == "pdf" {
		body := "Report: " + name + "\n\n" + strings.Join(rows, "\n")
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", name))
		c.String(http.StatusOK, simplePDF(body))
		return
	}
	c.Header("Content-Type", "application/vnd.ms-excel; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xls", name))
	c.String(http.StatusOK, strings.Join(rows, "\n"))
}

func (h *ReportHandler) rows(name string) []string {
	switch name {
	case "sales":
		var invoices []models.SalesInvoice
		h.db.Order("created_at desc").Find(&invoices)
		rows := []string{"number\ttotal\tpaid\tdate"}
		for _, inv := range invoices {
			rows = append(rows, fmt.Sprintf("%s\t%.2f\t%.2f\t%s", inv.Number, inv.Total, inv.PaidCash, inv.CreatedAt.Format("2006-01-02")))
		}
		return rows
	case "purchases":
		var invoices []models.PurchaseInvoice
		h.db.Order("created_at desc").Find(&invoices)
		rows := []string{"number\ttotal\tpaid\tdate"}
		for _, inv := range invoices {
			rows = append(rows, fmt.Sprintf("%s\t%.2f\t%.2f\t%s", inv.Number, inv.Total, inv.PaidCash, inv.CreatedAt.Format("2006-01-02")))
		}
		return rows
	case "stock":
		var movements []models.StockMovement
		h.db.Preload("Item").Order("created_at desc").Limit(500).Find(&movements)
		rows := []string{"item\ttype\tquantity\treference\tdate"}
		for _, m := range movements {
			rows = append(rows, fmt.Sprintf("%s\t%s\t%.3f\t%s\t%s", m.Item.Name, m.Type, m.Quantity, m.Reference, m.CreatedAt.Format("2006-01-02")))
		}
		return rows
	case "profits":
		var sales, purchases float64
		h.db.Model(&models.SalesInvoice{}).Select("COALESCE(SUM(total),0)").Scan(&sales)
		h.db.Model(&models.PurchaseInvoice{}).Select("COALESCE(SUM(total),0)").Scan(&purchases)
		return []string{"metric\tamount", fmt.Sprintf("sales\t%.2f", sales), fmt.Sprintf("purchases\t%.2f", purchases), fmt.Sprintf("gross_profit\t%.2f", sales-purchases)}
	default:
		var entries []models.JournalEntry
		h.db.Preload("Lines.Account").Order("created_at desc").Limit(500).Find(&entries)
		rows := []string{"entry\treference\taccount\tdebit\tcredit\tdate"}
		for _, e := range entries {
			for _, l := range e.Lines {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%.2f\t%.2f\t%s", e.Number, e.Reference, l.Account.Name, l.Debit, l.Credit, e.CreatedAt.Format("2006-01-02")))
			}
		}
		return rows
	}
}

func simplePDF(text string) string {
	escaped := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)", "\n", "\\n").Replace(text)
	stream := fmt.Sprintf("BT /F1 10 Tf 40 780 Td (%s) Tj ET", escaped)
	return fmt.Sprintf("%%PDF-1.4\n1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj\n4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n5 0 obj << /Length %d >> stream\n%s\nendstream endobj\ntrailer << /Root 1 0 R >>\n%%%%EOF", len(stream), stream)
}
