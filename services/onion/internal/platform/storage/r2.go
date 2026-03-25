package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

var (
	defaultClient     *R2Client
	ErrNotInitialized = errors.New("R2 storage not initialized")
	ErrUploadFailed   = errors.New("failed to upload to R2")
)

func Init(ctx context.Context) error {
	if !appConfig.IsR2Configured() {
		logger.Ctx(ctx).Warn("R2 not configured, file uploads will be disabled")

		return nil
	}

	accountID := appConfig.GetR2AccountID()
	accessKey := appConfig.GetR2AccessKeyID()
	secretKey := appConfig.GetR2SecretAccessKey()
	bucket := appConfig.GetR2Bucket()
	publicURL := appConfig.GetR2PublicURL()

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	defaultClient = &R2Client{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}

	logger.Ctx(ctx).Info("R2 storage initialized",
		logger.String("bucket", bucket),
		logger.String("public_url", publicURL),
	)

	return nil
}

func IsConfigured() bool {
	return defaultClient != nil
}

func Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	if defaultClient == nil {
		return "", ErrNotInitialized
	}

	_, err := defaultClient.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(defaultClient.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUploadFailed, err)
	}

	publicURL := fmt.Sprintf("%s/%s", defaultClient.publicURL, key)

	return publicURL, nil
}

func GenerateKey(prefix, filename string) string {
	timestamp := time.Now().Unix()

	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
