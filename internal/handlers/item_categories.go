package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ItemCategoryHandler struct{ service *services.ItemCategoryService }

func NewItemCategoryHandler(service *services.ItemCategoryService) *ItemCategoryHandler {
	return &ItemCategoryHandler{service: service}
}

func (h *ItemCategoryHandler) Index(c *gin.Context) {
	view := c.MustGet("view").(gin.H)
	view["categories"] = h.service.List(middleware.CurrentTenantID(c))
	c.HTML(http.StatusOK, "item_categories/index.html", view)
}

func (h *ItemCategoryHandler) Store(c *gin.Context) {
	if _, err := h.service.Create(middleware.CurrentTenantID(c), c.PostForm("name")); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/item-categories")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم حفظ التصنيف")
	c.Redirect(http.StatusFound, "/item-categories")
}

func (h *ItemCategoryHandler) Edit(c *gin.Context) {
	category, err := h.service.Find(parseUint(c.Param("id")))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	view := c.MustGet("view").(gin.H)
	view["category"] = category
	c.HTML(http.StatusOK, "item_categories/edit.html", view)
}

func (h *ItemCategoryHandler) Update(c *gin.Context) {
	if err := h.service.Update(middleware.CurrentTenantID(c), parseUint(c.Param("id")), c.PostForm("name")); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/item-categories/"+c.Param("id")+"/edit")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم تحديث التصنيف")
	c.Redirect(http.StatusFound, "/item-categories")
}

func (h *ItemCategoryHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(parseUint(c.Param("id"))); err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/item-categories")
		return
	}
	middleware.SetFlash(sessions.Default(c), "تم حذف التصنيف")
	c.Redirect(http.StatusFound, "/item-categories")
}

func (h *ItemCategoryHandler) QuickCreate(c *gin.Context) {
	category, err := h.service.Create(middleware.CurrentTenantID(c), c.PostForm("name"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": category.ID, "name": category.Name})
}
