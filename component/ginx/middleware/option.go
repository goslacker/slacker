package middleware

import "github.com/gin-gonic/gin"

func Options(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
