package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeffotoni/quick/pkg/glog"
)

var Logger = glog.New(
	glog.Config{
		Format: "json",
		Level:  glog.DEBUG,
	},
)

// HandleError is a helper function to handle an error
func HandleError(c *gin.Context, status int, message string, err error) {
	Logger.Error().Time().Err("error", err).Send()
	c.SecureJSON(
		status, gin.H{
			"message": message,
			"success": false,
		},
	)
}

// HandleSuccess is a helper function to handle a success
func HandleSuccess(c *gin.Context, message string) {
	c.SecureJSON(
		http.StatusOK, gin.H{
			"message": message,
			"success": true,
		},
	)
}

// HandleSuccessData is a helper function to handle a success and data
func HandleSuccessData(c *gin.Context, message string, data any) {
	c.SecureJSON(
		http.StatusOK, gin.H{
			"message": message,
			"success": true,
			"data":    data,
		},
	)
}
