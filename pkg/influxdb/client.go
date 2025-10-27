package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Client represents an InfluxDB client
type Client struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string
}

// NewClient creates a new InfluxDB client
func NewClient(serverURL, token, org, bucket string) (*Client, error) {
	client := influxdb2.NewClient(serverURL, token)

	// Test connection
	health, err := client.Health(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	if health.Status != "pass" {
		return nil, fmt.Errorf("InfluxDB health check failed: %s", health.Status)
	}

	writeAPI := client.WriteAPIBlocking(org, bucket)

	return &Client{
		client:   client,
		writeAPI: writeAPI,
		org:      org,
		bucket:   bucket,
	}, nil
}

// Close closes the InfluxDB connection
func (c *Client) Close() {
	c.client.Close()
}

// WritePoint writes a data point to InfluxDB
func (c *Client) WritePoint(measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error {
	point := write.NewPoint(measurement, tags, fields, timestamp)
	return c.writeAPI.WritePoint(context.Background(), point)
}

// WriteMetric writes a metric with tags and fields
func (c *Client) WriteMetric(measurement string, tags map[string]string, value interface{}, timestamp time.Time) error {
	fields := map[string]interface{}{
		"value": value,
	}
	return c.WritePoint(measurement, tags, fields, timestamp)
}

// WriteCounter writes a counter metric
func (c *Client) WriteCounter(name string, tags map[string]string, value int64) error {
	return c.WriteMetric(name, tags, value, time.Now())
}

// WriteTiming writes a timing metric
func (c *Client) WriteTiming(name string, tags map[string]string, duration time.Duration) error {
	return c.WriteMetric(name, tags, duration.Milliseconds(), time.Now())
}

// WriteGauge writes a gauge metric
func (c *Client) WriteGauge(name string, tags map[string]string, value float64) error {
	return c.WriteMetric(name, tags, value, time.Now())
}
