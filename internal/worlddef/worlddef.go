package worlddef

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Config represents worlddef configuration
type Config struct {
	WorldName        string
	WorldDescription string
	CoordinatorURL   string
	IdentityURL      string
	ProfileURL       string

	Log *logrus.Logger
}

// Register registers api routes
func Register(config *Config, router gin.IRouter) error {
	worlddef := map[string]interface{}{
		"name":        config.WorldName,
		"description": config.WorldDescription,
		"communication": map[string]interface{}{
			"url": config.CoordinatorURL,
		},
		"identity": map[string]interface{}{
			"url": config.IdentityURL,
		},
		"profile": map[string]interface{}{
			"url": config.ProfileURL,
		},
	}

	router.GET("/description", func(c *gin.Context) {
		c.JSON(200, worlddef)
	})

	return nil
}
