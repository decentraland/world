package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	logrus "github.com/sirupsen/logrus"
)

type MetricsClient struct {
	client *statsd.Client
	log    *logrus.Logger
}

func NewMetricsClient(appName string, log *logrus.Logger) (*MetricsClient, error) {
	client, err := statsd.New("", statsd.WithNamespace(fmt.Sprintf("%s.", appName)))
	if err != nil {
		return nil, err
	}

	c := &MetricsClient{
		client: client,
		log:    log,
	}

	return c, nil
}

func (c *MetricsClient) ServiceUp(name string) {
	c.simpleServiceCheck(name, statsd.Ok)
}

func (c *MetricsClient) GaugeInt(metric string, value int, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *MetricsClient) GaugeUint32(metric string, value uint32, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *MetricsClient) GaugeUint64(metric string, value uint64, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *MetricsClient) Close() {
	if err := c.client.Flush(); err != nil {
		c.log.WithError(err).Error("error flushing DD client")
	}
	if err := c.client.Close(); err != nil {
		c.log.WithError(err).Error("error closing DD client")
	}
}

func (c *MetricsClient) gauge(metric string, value float64, tags []string) {
	if err := c.client.Gauge(metric, float64(value), tags, 1); err != nil {
		c.log.WithError(err).Errorf("error sending metric %s", metric)
	}
}
func (c *MetricsClient) simpleServiceCheck(name string, status statsd.ServiceCheckStatus) {
	if err := c.client.SimpleServiceCheck(name, status); err != nil {
		c.log.WithError(err).Errorf("error sending service check %s", name)
	}
}
