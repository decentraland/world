package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// StatusResponse contains all the service dependencies status
type statusResponse struct {
	Ok     bool              `json:"ok"`
	Errors map[string]string `json:"errors"`
}

// runChecks checks the status of the service and returns a map of errors
func ServiceStatusHandler(runChecks func() map[string]string) func(c *gin.Context) {
	return func(c *gin.Context) {
		errors := runChecks()

		statusResponse := &statusResponse{Ok: len(errors) == 0, Errors: errors}

		if statusResponse.Ok {
			c.JSON(http.StatusOK, statusResponse)
		} else {
			c.JSON(http.StatusServiceUnavailable, statusResponse)
		}
	}
}
