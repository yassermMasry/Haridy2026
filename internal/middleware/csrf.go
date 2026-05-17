package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const csrfKey = "csrf_token"

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}
		session := sessions.Default(c)
		token, _ := session.Get(csrfKey).(string)
		if token == "" {
			token = randomToken()
			session.Set(csrfKey, token)
			_ = session.Save()
		}
		c.Set("csrf", token)

		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			if c.PostForm("_csrf") != token && c.GetHeader("X-CSRF-Token") != token {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		c.Next()
	}
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
