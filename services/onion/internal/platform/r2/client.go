package r2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrNotConfigured  = errors.New("R2 not configured")
	ErrUploadFailed   = errors.New("failed to upload file")
	ErrDownloadFailed = errors.New("failed to download file")
	ErrDeleteFailed   = errors.New("failed to delete file")
	ErrMissingConfig  = errors.New("missing required R2 configuration")
)

type Client struct {
	s3Client *s3.Client
	logger   *logger.Logger
}

type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type FileInfo struct {
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}

func NewClient(ctx context.Context) (*Client, error) {
	if !config.IsR2Configured() {
		return nil, ErrNotConfigured
	}

	accountID := config.GetR2AccountID()
	accessKeyID := config.GetR2AccessKeyID()
	secretAccessKey := config.GetR2SecretAccessKey()
	region := config.GetR2Region()

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil, ErrMissingConfig
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &Client{
		s3Client: s3Client,
		logger:   logger.Global(),
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, bucket, key string, reader io.Reader, mimeType string) (*UploadResult, error) {
	if c.s3Client == nil {
		return nil, ErrNotConfigured
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(mimeType),
	}

	_, err = c.s3Client.PutObject(ctx, input)
	if err != nil {
		c.logger.Error("failed to upload file to R2", logger.String("bucket", bucket), logger.String("key", key), logger.Error(err))

		return nil, fmt.Errorf("%w: %w", ErrUploadFailed, err)
	}

	c.logger.Info("file uploaded to R2", logger.String("bucket", bucket), logger.String("key", key), logger.Int("size", len(body)))

	return &UploadResult{
		URL:      c.GetPublicURL(bucket, key),
		Key:      key,
		Size:     int64(len(body)),
		MimeType: mimeType,
	}, nil
}

func (c *Client) DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if c.s3Client == nil {
		return nil, ErrNotConfigured
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		c.logger.Error("failed to download file from R2", logger.String("bucket", bucket), logger.String("key", key), logger.Error(err))

		return nil, fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	c.logger.Info("file downloaded from R2", logger.String("bucket", bucket), logger.String("key", key))

	return result.Body, nil
}

func (c *Client) DeleteFile(ctx context.Context, bucket, key string) error {
	if c.s3Client == nil {
		return ErrNotConfigured
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.DeleteObject(ctx, input)
	if err != nil {
		c.logger.Error("failed to delete file from R2", logger.String("bucket", bucket), logger.String("key", key), logger.Error(err))

		return fmt.Errorf("%w: %w", ErrDeleteFailed, err)
	}

	c.logger.Info("file deleted from R2", logger.String("bucket", bucket), logger.String("key", key))

	return nil
}

func (c *Client) GetFileInfo(ctx context.Context, bucket, key string) (*FileInfo, error) {
	if c.s3Client == nil {
		return nil, ErrNotConfigured
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		c.logger.Error("failed to get file info from R2", logger.String("bucket", bucket), logger.String("key", key), logger.Error(err))

		return nil, err
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &FileInfo{
		Key:      key,
		Size:     size,
		MimeType: aws.ToString(result.ContentType),
		URL:      c.GetPublicURL(bucket, key),
	}, nil
}

func (c *Client) GetPublicURL(bucket, key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	accountID := config.GetR2AccountID()

	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", accountID, bucket, key)
}

func (c *Client) IsConfigured() bool {
	return c.s3Client != nil
}

func (c *Client) Close() error {
	return nil
}
