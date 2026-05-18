package handlers

import (
	"net/http"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct{ service *services.AuthService }

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/login.html", c.MustGet("view"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	user, err := h.service.LoginWithContext(c.PostForm("username"), c.PostForm("password"), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.SetFlash(sessions.Default(c), err.Error())
		c.Redirect(http.StatusFound, "/login")
		return
	}
	session := sessions.Default(c)
	session.Set(middleware.UserIDKey, int(user.ID))
	session.Set(middleware.UserRoleKey, string(user.Role))
	if user.CurrentBranchID != nil {
		session.Set("current_branch_id", int(*user.CurrentBranchID))
	}
	_ = session.Save()
	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	c.Redirect(http.StatusFound, "/login")
}
