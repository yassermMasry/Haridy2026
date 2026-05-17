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

func (h *ItemHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["item"] = models.Item{}
	view["categories"] = h.service.Categories()
	view["action"] = "/items"
	c.HTML(http.StatusOK, "items/form.html", view)
}

func (h *ItemHandler) Store(c *gin.Context) {
	item := h.bindItem(c)
	if err := h.service.Save(&item); err != nil {
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
	view["categories"] = h.service.Categories()
	view["action"] = "/items/" + c.Param("id")
	c.HTML(http.StatusOK, "items/form.html", view)
}

func (h *ItemHandler) Update(c *gin.Context) {
	item := h.bindItem(c)
	item.ID = parseUint(c.Param("id"))
	if err := h.service.Save(&item); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/items/"+c.Param("id")+"/edit")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم تحديث الصنف")
	c.Redirect(http.StatusFound, "/items")
}

func (h *ItemHandler) Delete(c *gin.Context) {
	_ = h.service.Delete(parseUint(c.Param("id")))
	middleware.SetFlash(sessions.Default(c), "تم حذف الصنف")
	c.Redirect(http.StatusFound, "/items")
}

func (h *ItemHandler) bindItem(c *gin.Context) models.Item {
	categoryID := parseUint(c.PostForm("category_id"))
	var categoryPtr *uint
	if categoryID > 0 {
		categoryPtr = &categoryID
	}
	return models.Item{
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
