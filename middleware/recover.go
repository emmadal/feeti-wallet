package middleware

import (
	"fmt"
	"github.com/emmadal/feeti-wallet/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recover recovers from panics and returns a 500 Internal Server Error response.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				helpers.Logger.Error().Time().Err("panic", fmt.Errorf("%s", err)).Send()
				// Try to write header only if not already written
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(
						http.StatusInternalServerError, gin.H{
							"success": false,
							"message": "Internal server error",
						},
					)
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}
