package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	ddgin "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

type HttpMetricsConfig struct {
	TraceName            string
	AnalyticsRateEnabled bool
}

func EnableRouterMetrics(config *HttpMetricsConfig, router *gin.Engine) error {
	if len(config.TraceName) == 0 {
		return errors.New("missing metrics trace name")
	}

	router.Use(ddgin.Middleware(config.TraceName, ddgin.WithAnalytics(config.AnalyticsRateEnabled)))

	return nil
}
