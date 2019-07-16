package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/decentraland/world/internal/commons/logging"
)

type Client struct {
	client *statsd.Client
	log    *logging.Logger
}

func NewClient(appName string, log *logging.Logger) (*Client, error) {
	client, err := statsd.New("", statsd.WithNamespace(fmt.Sprintf("%s.", appName)))
	if err != nil {
		return nil, err
	}

	c := &Client{
		client: client,
		log:    log,
	}

	return c, nil
}

func (c *Client) ServiceUp(name string) {
	c.simpleServiceCheck(name, statsd.Ok)
}

func (c *Client) GaugeInt(metric string, value int, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *Client) GaugeUint32(metric string, value uint32, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *Client) GaugeUint64(metric string, value uint64, tags []string) {
	c.gauge(metric, float64(value), tags)
}

func (c *Client) Close() {
	if err := c.client.Flush(); err != nil {
		c.log.WithError(err).Error("error flushing DD client")
	}
	if err := c.client.Close(); err != nil {
		c.log.WithError(err).Error("error closing DD client")
	}
}

func (c *Client) gauge(metric string, value float64, tags []string) {
	if err := c.client.Gauge(metric, float64(value), tags, 1); err != nil {
		c.log.WithError(err).Errorf("error sending metric %s", metric)
	}
}
func (c *Client) simpleServiceCheck(name string, status statsd.ServiceCheckStatus) {
	if err := c.client.SimpleServiceCheck(name, status); err != nil {
		c.log.WithError(err).Errorf("error sending service check %s", name)
	}
}
