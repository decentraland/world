package version

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var version = "Not available"

type versionResponse struct {
	Version string `json:"version"`
}

func RegisterVersionEndpoint(r gin.IRoutes) {
	v := &versionResponse{Version: version}
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, v)
	})
}

func Version() string {
	return version
}
