package middleware

import (
	"strings"

	"haridy2026/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const TenantContextKey = "tenant_id"

func TenantResolver(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenant models.Tenant
		slug := c.GetHeader("X-Tenant")
		if slug == "" {
			host := c.Request.Host
			parts := strings.Split(host, ".")
			if len(parts) > 2 {
				slug = parts[0]
			}
		}
		if slug == "" {
			slug = "demo"
		}
		if err := db.Where("slug = ? OR subdomain = ? OR domain = ?", slug, slug, c.Request.Host).First(&tenant).Error; err == nil {
			c.Set(TenantContextKey, tenant.ID)
			c.Set("tenant", tenant)
		}
		c.Next()
	}
}

func CurrentTenantID(c *gin.Context) uint {
	if value, exists := c.Get(TenantContextKey); exists {
		if id, ok := value.(uint); ok {
			return id
		}
	}
	return 0
}
