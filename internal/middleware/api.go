package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func APIAuth(secrets ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenText := strings.TrimPrefix(header, "Bearer ")
		if tokenText == "" || tokenText == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		var token *jwt.Token
		var err error
		for _, secret := range secrets {
			token, err = jwt.Parse(tokenText, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, http.ErrAbortHandler
				}
				return []byte(secret), nil
			})
			if err == nil && token.Valid {
				break
			}
		}
		if token == nil || err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if id, ok := claims["user_id"].(float64); ok {
				c.Set("api_user_id", uint(id))
			}
		}
		c.Next()
	}
}
