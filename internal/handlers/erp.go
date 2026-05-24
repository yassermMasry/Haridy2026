package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ERPHandler struct {
	erp       *services.ERPService
	items     *services.ItemService
	customers *services.CustomerService
	suppliers *services.SupplierService
	returns   *services.ReturnService
}

func NewERPHandler(erp *services.ERPService, items *services.ItemService, customers *services.CustomerService, suppliers *services.SupplierService, returns *services.ReturnService) *ERPHandler {
	return &ERPHandler{erp: erp, items: items, customers: customers, suppliers: suppliers, returns: returns}
}

func (h *ERPHandler) Warehouses(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["warehouses"] = h.erp.Warehouses()
	view["branches"] = h.erp.Branches()
	view["items"] = h.items.List("", 1).Items
	c.HTML(http.StatusOK, "erp/warehouses.html", view)
}

func (h *ERPHandler) CreateWarehouse(c *gin.Context) {
	err := h.erp.CreateWarehouse(
		middleware.CurrentTenantID(c),
		parseUint(c.PostForm("branch_id")),
		c.PostForm("name"),
		c.PostForm("code"),
	)
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/warehouses")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم إضافة المخزن")
	c.Redirect(http.StatusFound, "/warehouses")
}

func (h *ERPHandler) Transfer(c *gin.Context) {
	err := h.erp.Transfer(parseUint(c.PostForm("item_id")), parseUint(c.PostForm("from_warehouse_id")), parseUint(c.PostForm("to_warehouse_id")), parseFloat(c.PostForm("quantity")), middleware.CurrentUserID(c))
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	}
	c.Redirect(http.StatusFound, "/warehouses")
}

func (h *ERPHandler) SetBranch(c *gin.Context) {
	middleware.SetCurrentBranch(c, parseUint(c.PostForm("branch_id")))
	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *ERPHandler) NewReceipt(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["customers"] = h.customers.List()
	c.HTML(http.StatusOK, "erp/receipt.html", view)
}

func (h *ERPHandler) CreateReceipt(c *gin.Context) {
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	voucher, err := h.erp.ReceiptVoucher(tenantPtr, parseUint(c.PostForm("customer_id")), parseFloat(c.PostForm("amount")), c.PostForm("description"), middleware.CurrentUserID(c))
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/vouchers/receipt/new")
		return
	}
	c.Redirect(http.StatusFound, "/vouchers/receipt/"+voucher.Number)
}

func (h *ERPHandler) PrintReceipt(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["number"] = c.Param("number")
	view["title"] = "Receipt Voucher"
	c.HTML(http.StatusOK, "erp/voucher_print.html", view)
}

func (h *ERPHandler) NewPayment(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["suppliers"] = h.suppliers.List()
	c.HTML(http.StatusOK, "erp/payment.html", view)
}

func (h *ERPHandler) CreatePayment(c *gin.Context) {
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	voucher, err := h.erp.PaymentVoucher(tenantPtr, parseUint(c.PostForm("supplier_id")), parseFloat(c.PostForm("amount")), c.PostForm("description"), middleware.CurrentUserID(c))
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/vouchers/payment/new")
		return
	}
	c.Redirect(http.StatusFound, "/vouchers/payment/"+voucher.Number)
}

func (h *ERPHandler) PrintPayment(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["number"] = c.Param("number")
	view["title"] = "Payment Voucher"
	c.HTML(http.StatusOK, "erp/voucher_print.html", view)
}

func (h *ERPHandler) SalesReturn(c *gin.Context) {
	if err := h.returns.SalesReturn(parseUint(c.PostForm("invoice_id")), c.PostForm("reason"), middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	}
	c.Redirect(http.StatusFound, "/sales")
}

func (h *ERPHandler) PurchaseReturn(c *gin.Context) {
	if err := h.returns.PurchaseReturn(parseUint(c.PostForm("invoice_id")), c.PostForm("reason"), middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	}
	c.Redirect(http.StatusFound, "/purchases")
}
