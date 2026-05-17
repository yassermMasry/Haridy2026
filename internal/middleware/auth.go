package middleware

import (
	"net/http"

	"haridy2026/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"

func CurrentUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	switch value := session.Get(UserIDKey).(type) {
	case uint:
		return value
	case int:
		return uint(value)
	case float64:
		return uint(value)
	default:
		return 0
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUserID(c) == 0 {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RoleRequired(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := sessions.Default(c).Get(UserRoleKey).(string)
		for _, allowed := range roles {
			if role == string(allowed) {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}
