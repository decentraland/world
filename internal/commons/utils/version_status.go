package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var version = "Not available"

type VersionResponse struct {
	Version string `json:"version"`
}

func RegisterVersionEndpoint(r gin.IRoutes) {
	v := &VersionResponse{Version: version}
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, v)
	})
}
