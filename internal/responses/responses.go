package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{"message": message, "data": data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func ServerError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
