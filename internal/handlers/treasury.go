package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type TreasuryHandler struct{ service *services.TreasuryService }

func NewTreasuryHandler(service *services.TreasuryService) *TreasuryHandler {
	return &TreasuryHandler{service: service}
}

func (h *TreasuryHandler) Index(c *gin.Context) {
	transactions, treasury := h.service.Transactions()
	view := c.MustGet("view").(gin.H)
	view["transactions"] = transactions
	view["treasury"] = treasury
	c.HTML(http.StatusOK, "treasury/index.html", view)
}

func (h *TreasuryHandler) Store(c *gin.Context) {
	err := h.service.Add(models.TreasuryTransactionType(c.PostForm("type")), parseFloat(c.PostForm("amount")), c.PostForm("description"), middleware.CurrentUserID(c))
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
	} else {
		middleware.SetFlash(sessions.Default(c), "تم تسجيل حركة الخزينة")
	}
	c.Redirect(http.StatusFound, "/treasury")
}
