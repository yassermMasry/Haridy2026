package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type CustomerHandler struct{ service *services.CustomerService }

func NewCustomerHandler(service *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["customers"] = h.service.List()
	c.HTML(http.StatusOK, "customers/index.html", view)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["customer"] = models.Customer{}
	view["action"] = "/customers"
	c.HTML(http.StatusOK, "customers/form.html", view)
}

func (h *CustomerHandler) Store(c *gin.Context) {
	customer := bindCustomer(c)
	if err := h.service.Save(&customer, middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/customers/new")
		return
	}
	c.Redirect(http.StatusFound, "/customers")
}

func (h *CustomerHandler) Edit(c *gin.Context) {
	customer, _, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["customer"] = customer
	view["action"] = "/customers/" + c.Param("id")
	c.HTML(http.StatusOK, "customers/form.html", view)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	customer := bindCustomer(c)
	customer.ID = parseUint(c.Param("id"))
	_ = h.service.Save(&customer, middleware.CurrentUserID(c))
	c.Redirect(http.StatusFound, "/customers")
}

func (h *CustomerHandler) Show(c *gin.Context) {
	customer, transactions, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["customer"] = customer
	view["transactions"] = transactions
	c.HTML(http.StatusOK, "customers/show.html", view)
}

func (h *CustomerHandler) Receive(c *gin.Context) {
	if err := h.service.Receive(parseUint(c.Param("id")), parseFloat(c.PostForm("amount")), middleware.CurrentUserID(c)); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	}
	c.Redirect(http.StatusFound, "/customers/"+c.Param("id"))
}

func (h *CustomerHandler) QuickCreate(c *gin.Context) {
	customer := bindCustomer(c)
	if err := h.service.Save(&customer, middleware.CurrentUserID(c)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": customer.ID, "name": customer.Name})
}

func bindCustomer(c *gin.Context) models.Customer {
	return models.Customer{Name: c.PostForm("name"), Phone: c.PostForm("phone"), Address: c.PostForm("address"), Notes: c.PostForm("notes"), CreditLimit: parseFloat(c.PostForm("credit_limit"))}
}
