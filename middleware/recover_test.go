package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emmadal/feeti-wallet/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.Recover())

	// Panic route
	router.GET(
		"/panic", func(c *gin.Context) {
			panic("Something broke!")
		},
	)

	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}
