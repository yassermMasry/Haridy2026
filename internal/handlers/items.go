package handlers

import (
	"net/http"
	"strconv"

	"haridy2026/internal/middleware"
	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ItemHandler struct{ service *services.ItemService }

func NewItemHandler(service *services.ItemService) *ItemHandler {
	return &ItemHandler{service: service}
}

func (h *ItemHandler) Index(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	view := c.MustGet("view").(gin.H)
	view["list"] = h.service.List(c.Query("q"), page)
	view["alerts"] = h.service.Alerts()
	c.HTML(http.StatusOK, "items/index.html", view)
}

func (h *ItemHandler) PriceList(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	tenantID := middleware.CurrentTenantID(c)
	view["warehouses"] = h.service.Warehouses(tenantID)
	view["priceList"] = h.service.PriceList(c.Query("q"), c.DefaultQuery("status", "all"), parseUint(c.Query("warehouse_id")))
	c.HTML(http.StatusOK, "items/price_list.html", view)
}

func (h *ItemHandler) UpdatePrice(c *gin.Context) {
	warnBelowCost, err := h.service.UpdateSalePrice(parseUint(c.Param("id")), parseFloat(c.PostForm("sale_price")))
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/price-list")
		return
	}
	if warnBelowCost {
		middleware.SetFlash(sessions.Default(c), "تم حفظ سعر البيع مع تحذير: السعر أقل من التكلفة")
	} else {
		middleware.SetFlash(sessions.Default(c), "تم تحديث سعر البيع")
	}
	c.Redirect(http.StatusFound, "/price-list")
}

func (h *ItemHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	tenantID := middleware.CurrentTenantID(c)
	view["item"] = models.Item{}
	view["categories"] = h.service.Categories(tenantID)
	view["warehouses"] = h.service.Warehouses(tenantID)
	view["selected_category_id"] = uint(0)
	view["is_create"] = true
	view["action"] = "/items"
	c.HTML(http.StatusOK, "items/form.html", view)
}

func (h *ItemHandler) Store(c *gin.Context) {
	item := h.bindItem(c)
	if err := h.service.CreateWithOpeningBalance(&item, parseUint(c.PostForm("warehouse_id"))); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/items/new")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم حفظ الصنف بنجاح")
	c.Redirect(http.StatusFound, "/items")
}

func (h *ItemHandler) Edit(c *gin.Context) {
	id := parseUint(c.Param("id"))
	item, err := h.service.Find(id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["item"] = item
	view["categories"] = h.service.Categories(middleware.CurrentTenantID(c))
	view["selected_category_id"] = uint(0)
	if item.CategoryID != nil {
		view["selected_category_id"] = *item.CategoryID
	}
	view["is_create"] = false
	view["action"] = "/items/" + c.Param("id")
	c.HTML(http.StatusOK, "items/form.html", view)
}

func (h *ItemHandler) Update(c *gin.Context) {
	item := h.bindItem(c)
	item.ID = parseUint(c.Param("id"))
	if err := h.service.SaveWithCategory(&item, ""); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/items/"+c.Param("id")+"/edit")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم تحديث الصنف")
	c.Redirect(http.StatusFound, "/items")
}

func (h *ItemHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(parseUint(c.Param("id"))); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/items")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم حذف الصنف")
	c.Redirect(http.StatusFound, "/items")
}

func (h *ItemHandler) bindItem(c *gin.Context) models.Item {
	categoryID := parseUint(c.PostForm("category_id"))
	var categoryPtr *uint
	if categoryID > 0 {
		categoryPtr = &categoryID
	}
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	return models.Item{
		TenantID:      tenantPtr,
		Name:          c.PostForm("name"),
		Code:          c.PostForm("code"),
		Barcode:       c.PostForm("barcode"),
		PurchasePrice: parseFloat(c.PostForm("purchase_price")),
		SalePrice:     parseFloat(c.PostForm("sale_price")),
		Quantity:      parseFloat(c.PostForm("quantity")),
		MinimumStock:  parseFloat(c.PostForm("minimum_stock")),
		CategoryID:    categoryPtr,
	}
}

func parseUint(value string) uint {
	n, _ := strconv.ParseUint(value, 10, 64)
	return uint(n)
}

func parseFloat(value string) float64 {
	n, _ := strconv.ParseFloat(value, 64)
	return n
}
