package middleware

import (
	"net/http"
	"sync"
	"time"

	"haridy2026/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self' https://cdn.tailwindcss.com https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline';")
		c.Next()
	}
}

type rateBucket struct {
	Count     int
	ResetTime time.Time
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateBucket{}
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		mu.Lock()
		bucket := buckets[key]
		if now.After(bucket.ResetTime) {
			bucket = rateBucket{ResetTime: now.Add(window)}
		}
		bucket.Count++
		buckets[key] = bucket
		allowed := bucket.Count <= limit
		mu.Unlock()
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func PermissionRequired(db *gorm.DB, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := CurrentUserID(c)
		if userID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var count int64
		db.Table("users").
			Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN role_permissions ON role_permissions.rbac_role_id = user_roles.rbac_role_id").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("users.id = ? AND permissions.code = ?", userID, code).
			Count(&count)
		if count == 0 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func CurrentBranchID(c *gin.Context) uint {
	session := sessions.Default(c)
	switch value := session.Get("current_branch_id").(type) {
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

func SetCurrentBranch(c *gin.Context, branchID uint) {
	session := sessions.Default(c)
	session.Set("current_branch_id", int(branchID))
	_ = session.Save()
}

func AuditRequest(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == http.MethodGet || c.Writer.Status() >= 400 {
			return
		}
		userID := CurrentUserID(c)
		var uid *uint
		if userID > 0 {
			uid = &userID
		}
		_ = db.Create(&models.AuditLog{UserID: uid, Action: c.Request.Method, Entity: "http_request", Details: c.FullPath()}).Error
	}
}
