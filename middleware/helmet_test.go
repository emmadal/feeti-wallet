package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHelmet(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Helmet())
	r.GET(
		"/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		},
	)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())

	// Security headers tests
	expectedHeaders := map[string]string{
		"X-Frame-Options":           "SAMEORIGIN",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self';",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
	}

	// Check that expected headers exist and have the correct values
	for key, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(key)
		assert.Equal(t, expectedValue, actualValue, "Header '%s' has unexpected value", key)
	}

	// X-Powered-By and Server headers should be empty
	assert.Empty(t, w.Header().Get("X-Powered-By"), "X-Powered-By header should be empty")
	assert.Empty(t, w.Header().Get("Server"), "Server header should be empty")
}

// TestHelmetDoesNotOverrideExistingHeaders tests that the Helmet middleware
// doesn't override headers that are already set in the request pipeline.
func TestHelmetDoesNotOverrideExistingHeaders(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a handler that sets a custom header before Helmet middleware
	r.Use(
		func(c *gin.Context) {
			c.Header("X-Frame-Options", "DENY") // Set a custom value
			c.Next()
		},
	)

	// Apply Helmet middleware
	r.Use(Helmet())

	r.GET(
		"/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		},
	)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// The custom value should be preserved
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

// TestHelmetAllHeaders performs a comprehensive test of all headers
// including their presence and correct values
func TestHelmetAllHeaders(t *testing.T) {
	// Common setup
	setupRequest := func() (*httptest.ResponseRecorder, *http.Request) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(Helmet())
		r.GET(
			"/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			},
		)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w, req
	}

	w, _ := setupRequest()

	// Test all headers individually for more precise error messages on failure
	t.Run(
		"X-Frame-Options", func(t *testing.T) {
			assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
		},
	)

	t.Run(
		"X-Content-Type-Options", func(t *testing.T) {
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		},
	)

	t.Run(
		"X-XSS-Protection", func(t *testing.T) {
			assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		},
	)

	t.Run(
		"Referrer-Policy", func(t *testing.T) {
			assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		},
	)

	t.Run(
		"Content-Security-Policy", func(t *testing.T) {
			expected := "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self';"
			assert.Equal(t, expected, w.Header().Get("Content-Security-Policy"))
		},
	)

	t.Run(
		"Strict-Transport-Security", func(t *testing.T) {
			expected := "max-age=63072000; includeSubDomains; preload"
			assert.Equal(t, expected, w.Header().Get("Strict-Transport-Security"))
		},
	)

	t.Run(
		"Permissions-Policy", func(t *testing.T) {
			expected := "geolocation=(), microphone=(), camera=()"
			assert.Equal(t, expected, w.Header().Get("Permissions-Policy"))
		},
	)

	t.Run(
		"X-Powered-By and Server headers are empty", func(t *testing.T) {
			assert.Empty(t, w.Header().Get("X-Powered-By"))
			assert.Empty(t, w.Header().Get("Server"))
		},
	)
}
