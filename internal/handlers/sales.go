package handlers

import (
	"net/http"
	"strconv"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SalesHandler struct {
	service   *services.SalesService
	items     *services.ItemService
	customers *services.CustomerService
}

func NewSalesHandler(service *services.SalesService, items *services.ItemService, customers *services.CustomerService) *SalesHandler {
	return &SalesHandler{service: service, items: items, customers: customers}
}

func (h *SalesHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["invoices"] = h.service.List()
	c.HTML(http.StatusOK, "sales/index.html", view)
}

func (h *SalesHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["items"] = h.items.List("", 1).Items
	view["customers"] = h.customers.List()
	c.HTML(http.StatusOK, "sales/create.html", view)
}

func (h *SalesHandler) Store(c *gin.Context) {
	_ = c.Request.ParseForm()
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	itemIDs := c.Request.PostForm["item_id[]"]
	quantities := c.Request.PostForm["quantity[]"]
	prices := c.Request.PostForm["price[]"]
	var lines []services.SaleLineInput
	for i, rawID := range itemIDs {
		itemID := parseUint(rawID)
		if itemID == 0 || i >= len(quantities) {
			continue
		}
		price := 0.0
		if i < len(prices) {
			price = parseFloat(prices[i])
		}
		lines = append(lines, services.SaleLineInput{ItemID: itemID, Quantity: parseFloat(quantities[i]), UnitPrice: price})
	}
	invoice, err := h.service.Create(services.SaleInput{
		TenantID: tenantPtr, UserID: middleware.CurrentUserID(c), CustomerID: parseUint(c.PostForm("customer_id")), PaymentType: c.PostForm("payment_type"), Discount: parseFloat(c.PostForm("discount")), Tax: parseFloat(c.PostForm("tax")), PaidCash: parseFloat(c.PostForm("paid_cash")), Lines: lines,
	})
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/sales/new")
		return
	}
	c.Redirect(http.StatusFound, "/sales/"+strconv.Itoa(int(invoice.ID)))
}

func (h *SalesHandler) Show(c *gin.Context) {
	invoice, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil && err == gorm.ErrRecordNotFound {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["invoice"] = invoice
	c.HTML(http.StatusOK, "sales/show.html", view)
}
