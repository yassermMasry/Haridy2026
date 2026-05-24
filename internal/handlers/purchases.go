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

type PurchaseHandler struct {
	service   *services.PurchaseService
	items     *services.ItemService
	suppliers *services.SupplierService
}

func NewPurchaseHandler(service *services.PurchaseService, items *services.ItemService, suppliers *services.SupplierService) *PurchaseHandler {
	return &PurchaseHandler{service: service, items: items, suppliers: suppliers}
}

func (h *PurchaseHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["invoices"] = h.service.List()
	c.HTML(http.StatusOK, "purchases/index.html", view)
}

func (h *PurchaseHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["items"] = h.items.List("", 1).Items
	view["suppliers"] = h.suppliers.List()
	c.HTML(http.StatusOK, "purchases/create.html", view)
}

func (h *PurchaseHandler) Store(c *gin.Context) {
	_ = c.Request.ParseForm()
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	itemIDs := c.Request.PostForm["item_id[]"]
	quantities := c.Request.PostForm["quantity[]"]
	costs := c.Request.PostForm["cost[]"]
	var lines []services.PurchaseLineInput
	for i, rawID := range itemIDs {
		if i >= len(quantities) || i >= len(costs) {
			continue
		}
		lines = append(lines, services.PurchaseLineInput{ItemID: parseUint(rawID), Quantity: parseFloat(quantities[i]), UnitCost: parseFloat(costs[i])})
	}
	invoice, err := h.service.Create(services.PurchaseInput{TenantID: tenantPtr, UserID: middleware.CurrentUserID(c), SupplierID: parseUint(c.PostForm("supplier_id")), PaymentType: c.PostForm("payment_type"), Discount: parseFloat(c.PostForm("discount")), Tax: parseFloat(c.PostForm("tax")), PaidCash: parseFloat(c.PostForm("paid_cash")), Lines: lines})
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/purchases/new")
		return
	}
	c.Redirect(http.StatusFound, "/purchases/"+strconv.Itoa(int(invoice.ID)))
}

func (h *PurchaseHandler) Show(c *gin.Context) {
	invoice, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil && err == gorm.ErrRecordNotFound {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["invoice"] = invoice
	c.HTML(http.StatusOK, "purchases/show.html", view)
}
