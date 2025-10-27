package statsd

import (
	"fmt"
	"net"
	"time"
)

// Client represents a StatsD client
type Client struct {
	conn   net.Conn
	prefix string
}

// NewClient creates a new StatsD client
func NewClient(addr, prefix string) (*Client, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to StatsD: %w", err)
	}

	return &Client{
		conn:   conn,
		prefix: prefix,
	}, nil
}

// Close closes the StatsD connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Incr increments a counter metric
func (c *Client) Incr(name string, tags map[string]string) error {
	metric := c.buildMetric(name, tags)
	_, err := fmt.Fprintf(c.conn, "%s:1|c\n", metric)
	return err
}

// Timing records a timing metric
func (c *Client) Timing(name string, duration time.Duration, tags map[string]string) error {
	metric := c.buildMetric(name, tags)
	_, err := fmt.Fprintf(c.conn, "%s:%d|ms\n", metric, duration.Milliseconds())
	return err
}

// Gauge sets a gauge metric
func (c *Client) Gauge(name string, value float64, tags map[string]string) error {
	metric := c.buildMetric(name, tags)
	_, err := fmt.Fprintf(c.conn, "%s:%f|g\n", metric, value)
	return err
}

// buildMetric constructs the metric name with tags
func (c *Client) buildMetric(name string, tags map[string]string) string {
	metric := name
	if c.prefix != "" {
		metric = fmt.Sprintf("%s.%s", c.prefix, name)
	}

	// Add tags in StatsD format
	if len(tags) > 0 {
		for key, value := range tags {
			metric = fmt.Sprintf("%s;%s=%s", metric, key, value)
		}
	}

	return metric
}
