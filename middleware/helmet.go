package middleware

import (
	"github.com/gin-gonic/gin"
)

// Helmet is a middleware function that sets various security headers.
func Helmet() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Helper to set header only if not already set
		setHeaderIfNotExist := func(key, value string) {
			if c.Writer.Header().Get(key) == "" {
				c.Header(key, value)
			}
		}

		// Set security headers if they don't already exist
		setHeaderIfNotExist("X-Frame-Options", "SAMEORIGIN")
		setHeaderIfNotExist("X-Content-Type-Options", "nosniff")
		setHeaderIfNotExist("X-XSS-Protection", "1; mode=block")
		setHeaderIfNotExist("Referrer-Policy", "strict-origin-when-cross-origin")
		setHeaderIfNotExist(
			"Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self';",
		)
		setHeaderIfNotExist("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		setHeaderIfNotExist("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Always remove these headers by setting them to empty
		c.Header("X-Powered-By", "")
		c.Header("Server", "")

		c.Next()
	}
}
