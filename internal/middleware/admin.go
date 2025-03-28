package middleware

import (
	"github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement admin check
		c.Next()
	}
}
