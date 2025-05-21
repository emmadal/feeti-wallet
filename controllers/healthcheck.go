package controllers

import (
	status "github.com/emmadal/feeti-module/status"
	"github.com/gin-gonic/gin"
	"time"
)

// HealthCheck check is a health check endpoint for kubernetes
func HealthCheck(c *gin.Context) {
	status.HandleSuccessData(
		c, "OK", gin.H{
			"status":  "up",
			"time":    time.Now().Format(time.RFC3339),
			"service": "Wallet Service",
		},
	)
}
