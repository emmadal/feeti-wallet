package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

const defaultTimeout = 5 * time.Second

// Timeout middleware
func Timeout(duration time.Duration) gin.HandlerFunc {
	if duration <= 0 {
		duration = defaultTimeout
	}
	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithHandler(
			func(c *gin.Context) {
				c.Next()
			},
		),
		timeout.WithResponse(
			func(c *gin.Context) {
				c.SecureJSON(
					http.StatusRequestTimeout, gin.H{
						"message": "Request timed out",
					},
				)
			},
		),
	)
}
