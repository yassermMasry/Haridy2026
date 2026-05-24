package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ActivationHandler struct {
	service *services.ActivationService
}

func NewActivationHandler(service *services.ActivationService) *ActivationHandler {
	return &ActivationHandler{service: service}
}

func (h *ActivationHandler) Show(c *gin.Context) {
	c.HTML(http.StatusOK, "activation/index.html", c.MustGet("view"))
}

func (h *ActivationHandler) Activate(c *gin.Context) {
	if err := h.service.Activate(middleware.CurrentTenantID(c), c.PostForm("code")); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/activation")
		return
	}
	middleware.SetFlash(sessions.Default(c), services.ActivationSuccessMessage)
	c.Redirect(http.StatusFound, "/activation")
}
