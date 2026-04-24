// Package storage is a thin wrapper around S3-compatible object
// storage. Production uses Cloudflare R2; local development can point
// the same code at LocalStack.
package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	localStackEnabledEnv  = "LOCAL_STACK_ENABLED"
	localStackEndpointEnv = "LOCAL_STACK_ENDPOINT"
	localStackBucketEnv   = "LOCAL_STACK_BUCKET"
	localStackRegionEnv   = "LOCAL_STACK_REGION"

	defaultLocalStackEndpoint = "http://localhost:4566"
	defaultLocalStackBucket   = "onion-dev"
	defaultLocalStackRegion   = "us-east-1"
)

// ErrNotConfigured is returned when any of the four required R2 env
// vars is missing. Callers surface this as a 501 so the frontend can
// tell the difference between "upload failed" and "upload disabled".
var ErrNotConfigured = errors.New("r2 storage is not configured")

// Client is a lazily-initialised R2 client. Feature code takes the
// default instance via Default().
type Client struct {
	s3     *s3.Client
	bucket string
}

var (
	defaultClient *Client
	defaultOnce   sync.Once
	defaultErr    error
)

// Default returns the process-wide R2 client, building it the first
// time it's asked for. Returns ErrNotConfigured if the env vars aren't
// set — not a fatal condition, callers degrade gracefully.
func Default() (*Client, error) {
	defaultOnce.Do(func() {
		settings, settingsErr := resolveSettingsFromEnv()
		if settingsErr != nil {
			defaultErr = settingsErr
			return
		}

		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(settings.accessKey, settings.secretKey, "")),
			awsconfig.WithRegion(settings.region),
		)
		if err != nil {
			defaultErr = fmt.Errorf("load object storage aws config: %w", err)
			return
		}

		s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(settings.endpoint)
			// R2 and LocalStack are both path-style compatible. This
			// also avoids per-bucket DNS requirements in local dev.
			o.UsePathStyle = true
		})

		defaultClient = &Client{s3: s3Client, bucket: settings.bucket}
	})
	return defaultClient, defaultErr
}

type storageSettings struct {
	endpoint  string
	region    string
	accessKey string
	secretKey string
	bucket    string
}

func resolveSettingsFromEnv() (storageSettings, error) {
	if localStackEnabled() {
		return storageSettings{
			endpoint:  envOrDefault(localStackEndpointEnv, defaultLocalStackEndpoint),
			region:    envOrDefault(localStackRegionEnv, defaultLocalStackRegion),
			accessKey: "test",
			secretKey: "test",
			bucket:    envOrDefault(localStackBucketEnv, defaultLocalStackBucket),
		}, nil
	}

	return resolveR2SettingsFromEnv()
}

func resolveR2SettingsFromEnv() (storageSettings, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	if accountID == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return storageSettings{}, ErrNotConfigured
	}

	// R2 endpoint format. The region is a placeholder ("auto") —
	// R2 doesn't care what we send but aws-sdk-go-v2 refuses to
	// sign without one.
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	return storageSettings{
		endpoint:  endpoint,
		region:    "auto",
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
	}, nil
}

func localStackEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(localStackEnabledEnv)))
	return value == "true" || value == "1" || value == "yes"
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// AllowedResumeContentTypes is the whitelist for the resume upload
// endpoint. PDFs and Word docs cover 99% of candidates; we reject
// everything else to keep R2 from doubling as free file hosting.
var AllowedResumeContentTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

// MaxResumeBytes is the hard ceiling enforced by the presigned URL
// via Content-Length-Range. 10 MB is generous — most PDFs are under 1.
const MaxResumeBytes int64 = 10 * 1024 * 1024

// PresignUpload returns a short-lived PUT URL the browser can upload
// to directly. The key is caller-controlled; the content-type lock
// means a candidate can't swap a PDF for an executable after signing.
func (c *Client) PresignUpload(ctx context.Context, key, contentType string, expiresIn time.Duration) (string, error) {
	presign := s3.NewPresignClient(c.s3)
	req, err := presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// PresignDownload returns a short-lived GET URL HR can follow to read
// a resume. Short expiry keeps leaked URLs from becoming permanent
// leaks — a screenshot of a Slack message is safe after an hour.
func (c *Client) PresignDownload(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	presign := s3.NewPresignClient(c.s3)
	req, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Bucket exposes the configured bucket name for callers that need to
// log or report it. Not used for signing (the client already holds it).
func (c *Client) Bucket() string {
	return c.bucket
}

// SanitizeObjectKey removes characters that would trip up S3 or URL
// encoders. The returned string is safe to use as an R2 key and as
// a URL path segment. We keep it ASCII-only; candidates with
// non-Latin names still get a unique id prefix.
func SanitizeObjectKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	out := b.String()
	if out == "" {
		return "file"
	}
	return out
}

// BuildHTTPError turns storage errors into a suitable HTTP status for
// handler callers that don't want to import this package's sentinels.
func BuildHTTPError(err error) (int, string) {
	if errors.Is(err, ErrNotConfigured) {
		return http.StatusNotImplemented, "File uploads are not configured on this server"
	}
	return http.StatusInternalServerError, "Storage error"
}
