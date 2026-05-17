package middleware

import (
	"html/template"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ViewData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		c.Set("view", gin.H{
			"csrf":        c.GetString("csrf"),
			"flash":       PopFlash(session),
			"currentPath": c.Request.URL.Path,
			"appName":     "Haridy Inventory",
			"safe": func(s string) template.HTML {
				return template.HTML(s)
			},
		})
		c.Next()
	}
}
