package gindcl

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func PrefligthChecksMiddleware(allowedMethods string, allowedHeaders string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.AbortWithStatus(http.StatusNoContent)
	}
}
