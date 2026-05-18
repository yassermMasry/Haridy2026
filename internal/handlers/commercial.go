package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CommercialHandler struct{ service *services.CommercialService }

func NewCommercialHandler(service *services.CommercialService) *CommercialHandler {
	return &CommercialHandler{service: service}
}

func (h *CommercialHandler) Landing(c *gin.Context) {
	c.HTML(http.StatusOK, "commercial/landing.html", gin.H{"csrf": c.GetString("csrf"), "appName": "Haridy ERP SaaS"})
}

func (h *CommercialHandler) Onboarding(c *gin.Context) {
	c.HTML(http.StatusOK, "commercial/onboarding.html", gin.H{"csrf": c.GetString("csrf"), "appName": "Haridy ERP SaaS"})
}

func (h *CommercialHandler) CreateCompany(c *gin.Context) {
	hash, err := bcrypt.GenerateFromPassword([]byte(c.PostForm("password")), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	tenant, err := h.service.OnboardCompany(c.PostForm("company_name"), c.PostForm("slug"), c.PostForm("username"), string(hash))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tenant": tenant.Slug, "login": "/login"})
}

func (h *CommercialHandler) TrialBalance(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.service.TrialBalance(middleware.CurrentTenantID(c))})
}

func (h *CommercialHandler) FinancialStatements(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.service.FinancialSummary(middleware.CurrentTenantID(c))})
}

func (h *CommercialHandler) EInvoice(c *gin.Context) {
	doc, err := h.service.GenerateEInvoice(middleware.CurrentTenantID(c), parseUint(c.Param("invoice_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": doc})
}

func (h *CommercialHandler) CloseYear(c *gin.Context) {
	err := h.service.CloseFiscalYear(middleware.CurrentTenantID(c), parseUint(c.PostForm("fiscal_year_id")), middleware.CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "closed"})
}
