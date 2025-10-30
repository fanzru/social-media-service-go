package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	// Register image format decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/fanzru/social-media-service-go/infrastructure/config"
	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/s3"
)

// ImageStorageService handles image upload and processing
type ImageStorageService struct {
	config   *config.StorageConfig
	s3Client *s3.Client
	logger   *logger.Logger
}

// NewImageStorageService creates a new image storage service
func NewImageStorageService(cfg *config.StorageConfig) *ImageStorageService {
	service := &ImageStorageService{
		config: cfg,
		logger: logger.GetGlobal(),
	}

	// Always initialize S3 client
	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		service.logger.Error("Failed to create S3 client", "error", err.Error())
		panic(fmt.Sprintf("S3 client initialization failed: %v", err))
	}
	service.s3Client = s3Client
	service.logger.Info("S3 client initialized", "bucket", cfg.S3Bucket, "region", cfg.S3Region)

	return service
}

// ProcessAndUploadImage processes and uploads an image directly to S3
func (s *ImageStorageService) ProcessAndUploadImage(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	// Validate file
	if err := s.validateFile(header); err != nil {
		return "", "", fmt.Errorf("file validation failed: %w", err)
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate a stable timestamp-based base name
	timestamp := time.Now().UnixNano()

	// Upload original file in its original format
	originalExt := strings.ToLower(filepath.Ext(header.Filename))
	if originalExt == "" {
		originalExt = ".bin"
	}
	originalKey := fmt.Sprintf("post_%d_orig%s", timestamp, originalExt)
	contentType := contentTypeFromExt(originalExt)
	if err := s.s3Client.Upload(context.Background(), originalKey, bytes.NewReader(fileContent), contentType); err != nil {
		return "", "", fmt.Errorf("original image upload failed: %w", err)
	}
	// Process image (resize and convert to JPG)
	processedImage, err := s.processImage(fileContent)
	if err != nil {
		return "", "", fmt.Errorf("image processing failed: %w", err)
	}

	// Generate processed filename (always .jpg)
	processedKey := fmt.Sprintf("post_%d.jpg", timestamp)

	// Upload processed image directly to S3
	imagePath, imageURL, err := s.uploadToS3(processedImage, processedKey)
	if err != nil {
		return "", "", fmt.Errorf("image upload failed: %w", err)
	}

	return imagePath, imageURL, nil
}

// validateFile validates the uploaded file
func (s *ImageStorageService) validateFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > s.config.MaxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.config.MaxSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	for _, allowedExt := range s.config.AllowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("file extension %s is not allowed. Allowed extensions: %v", ext, s.config.AllowedExts)
}

// processImage processes the image (resize and convert to JPG)
func (s *ImageStorageService) processImage(imageData []byte) ([]byte, error) {
	// Decode image
	img, err := imaging.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image
	resizedImg := imaging.Resize(img, s.config.ImageResizeWidth, s.config.ImageResizeHeight, imaging.Lanczos)

	// Encode as JPEG
	var buf bytes.Buffer
	err = imaging.Encode(&buf, resizedImg, imaging.JPEG, imaging.JPEGQuality(s.config.ImageQuality))
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// generateFilename generates a unique filename
func (s *ImageStorageService) generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	if ext != ".jpg" && ext != ".jpeg" {
		ext = ".jpg"
	}

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("post_%d%s", timestamp, ext)
}

// uploadToS3 uploads image to S3
func (s *ImageStorageService) uploadToS3(imageData []byte, filename string) (string, string, error) {
	ctx := context.Background()

	// Upload to S3 using our wrapper
	err := s.s3Client.Upload(ctx, filename, bytes.NewReader(imageData), "image/jpeg")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate URLs
	imagePath := filename
	imageURL := s.s3Client.GetURL(filename)

	s.logger.Info("Image uploaded to S3", "filename", filename, "bucket", s.config.S3Bucket)

	return imagePath, imageURL, nil
}

// DeleteImage deletes an image from S3
func (s *ImageStorageService) DeleteImage(imagePath string) error {
	// Delete processed image
	_ = s.deleteFromS3(imagePath)

	// Also attempt to delete any plausible original variant derived from the processed key
	base := strings.TrimSuffix(imagePath, filepath.Ext(imagePath))
	candidates := []string{
		base + "_orig.png",
		base + "_orig.jpg",
		base + "_orig.jpeg",
		base + "_orig.bmp",
	}
	for _, key := range candidates {
		_ = s.deleteFromS3(key)
	}
	return nil
}

// deleteFromS3 deletes image from S3
func (s *ImageStorageService) deleteFromS3(imagePath string) error {
	ctx := context.Background()

	err := s.s3Client.Delete(ctx, imagePath)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	s.logger.Info("Image deleted from S3", "path", imagePath)
	return nil
}

// GenerateImageURL generates the public URL for an image from S3
func (s *ImageStorageService) GenerateImageURL(filename string) string {
	return s.s3Client.GetURL(filename)
}

// contentTypeFromExt maps a file extension to an image content type.
func contentTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}
