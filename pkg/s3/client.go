package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fanzru/social-media-service-go/infrastructure/config"
	"github.com/fanzru/social-media-service-go/pkg/logger"
)

// Client wraps AWS S3 client with simplified interface
type Client struct {
	client  *s3.Client
	bucket  string
	region  string
	baseURL string
	logger  *logger.Logger
}

// NewClient creates a new S3 client
func NewClient(cfg *config.StorageConfig) (*Client, error) {
	// Validate required S3 configuration
	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		return nil, fmt.Errorf("S3 credentials are required: S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY must be set")
	}
	if cfg.S3Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required: S3_BUCKET must be set")
	}

	// Create AWS config manually to avoid shared config issues
	awsConfig := aws.Config{
		Region: cfg.S3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKeyID,
			cfg.S3SecretAccessKey,
			"",
		),
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		}
	})

	return &Client{
		client:  s3Client,
		bucket:  cfg.S3Bucket,
		region:  cfg.S3Region,
		baseURL: cfg.S3ImageBaseURL,
		logger:  logger.GetGlobal(),
	}, nil
}

// Upload uploads data to S3
func (c *Client) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
		// Remove ACL for Cloudflare R2 compatibility
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	c.logger.Info("File uploaded to S3", "key", key, "bucket", c.bucket)
	return nil
}

// Delete deletes an object from S3
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	c.logger.Info("File deleted from S3", "key", key, "bucket", c.bucket)
	return nil
}

// GetURL generates the public URL for an object
func (c *Client) GetURL(key string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, key)
}

// Exists checks if an object exists in S3
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error by checking the error message
		// This is a simpler approach for AWS SDK v2
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if object exists: %w", err)
	}
	return true, nil
}

// ListObjects lists objects in the bucket with a prefix
func (c *Client) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(c.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(maxKeys),
	}

	result, err := c.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var keys []string
	for _, obj := range result.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

// GetObject retrieves an object from S3
func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

// CopyObject copies an object within S3
func (c *Client) CopyObject(ctx context.Context, sourceKey, destKey string) error {
	source := fmt.Sprintf("%s/%s", c.bucket, sourceKey)
	_, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(c.bucket),
		CopySource: aws.String(source),
		Key:        aws.String(destKey),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object in S3: %w", err)
	}

	c.logger.Info("Object copied in S3", "source", sourceKey, "dest", destKey)
	return nil
}
