package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SupplierHandler struct{ service *services.SupplierService }

func NewSupplierHandler(service *services.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: service}
}

func (h *SupplierHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["suppliers"] = h.service.List()
	c.HTML(http.StatusOK, "suppliers/index.html", view)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["supplier"] = models.Supplier{}
	view["action"] = "/suppliers"
	c.HTML(http.StatusOK, "suppliers/form.html", view)
}

func (h *SupplierHandler) Store(c *gin.Context) {
	supplier := bindSupplier(c)
	if err := h.service.Save(&supplier, middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/suppliers/new")
		return
	}
	c.Redirect(http.StatusFound, "/suppliers")
}

func (h *SupplierHandler) Edit(c *gin.Context) {
	supplier, _, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["supplier"] = supplier
	view["action"] = "/suppliers/" + c.Param("id")
	c.HTML(http.StatusOK, "suppliers/form.html", view)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	supplier := bindSupplier(c)
	supplier.ID = parseUint(c.Param("id"))
	_ = h.service.Save(&supplier, middleware.CurrentUserID(c))
	c.Redirect(http.StatusFound, "/suppliers")
}

func (h *SupplierHandler) Show(c *gin.Context) {
	supplier, transactions, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["supplier"] = supplier
	view["transactions"] = transactions
	c.HTML(http.StatusOK, "suppliers/show.html", view)
}

func (h *SupplierHandler) Pay(c *gin.Context) {
	if err := h.service.Pay(parseUint(c.Param("id")), parseFloat(c.PostForm("amount")), middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	}
	c.Redirect(http.StatusFound, "/suppliers/"+c.Param("id"))
}

func bindSupplier(c *gin.Context) models.Supplier {
	return models.Supplier{Name: c.PostForm("name"), Phone: c.PostForm("phone"), Address: c.PostForm("address"), Notes: c.PostForm("notes")}
}
