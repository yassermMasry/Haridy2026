package handlers

import (
	"net/http"

	"haridy2026/internal/services"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{ service *services.DashboardService }

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["stats"] = h.service.Stats()
	c.HTML(http.StatusOK, "dashboard/index.html", view)
}
