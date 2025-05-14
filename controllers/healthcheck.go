package controllers

import (
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/gin-gonic/gin"
	"time"
)

// HealthCheck check is a health check endpoint for kubernetes
func HealthCheck(c *gin.Context) {
	helpers.HandleSuccessData(
		c, "OK", gin.H{
			"status": "up",
			"time":   time.Now().Format(time.RFC3339),
		},
	)
}
