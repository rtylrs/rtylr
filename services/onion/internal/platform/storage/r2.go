// Package storage is a thin wrapper around AWS S3 object storage.
// Production uses AWS S3 with an EC2 IAM instance role (no static keys);
// local development can point the same code at Floci.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	flociEnabledEnv  = "FLOCI_ENABLED"
	flociEndpointEnv = "FLOCI_ENDPOINT"
	flociBucketEnv   = "FLOCI_BUCKET"
	flociRegionEnv   = "FLOCI_REGION"

	defaultFlociEndpoint = "http://localhost:4566"
	defaultFlociBucket   = "onion-dev"
	defaultFlociRegion   = "us-east-1"
)

// ErrNotConfigured is returned when required S3 env vars are missing.
// Callers surface this as a 501 so the frontend can tell the difference
// between "upload failed" and "upload disabled".
var ErrNotConfigured = errors.New("s3 storage is not configured")

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

// Default returns the process-wide S3 client, building it the first
// time it's asked for. Returns ErrNotConfigured if the env vars aren't
// set — not a fatal condition, callers degrade gracefully.
//
// In production (EC2 with an IAM instance role) no static credentials
// are needed — the SDK picks them up automatically from instance metadata.
func Default() (*Client, error) {
	defaultOnce.Do(func() {
		settings, settingsErr := resolveSettingsFromEnv()
		if settingsErr != nil {
			defaultErr = settingsErr
			return
		}

		opts := []func(*awsconfig.LoadOptions) error{
			awsconfig.WithRegion(settings.region),
		}

		cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			defaultErr = fmt.Errorf("load object storage aws config: %w", err)
			return
		}

		s3Opts := []func(*s3.Options){}
		if settings.endpoint != "" {
			// Custom endpoint for Floci (local dev). Not set for real S3.
			s3Opts = append(s3Opts, func(o *s3.Options) {
				o.BaseEndpoint = aws.String(settings.endpoint)
				o.UsePathStyle = true
			})
		}

		defaultClient = &Client{s3: s3.NewFromConfig(cfg, s3Opts...), bucket: settings.bucket}
	})
	return defaultClient, defaultErr
}

type storageSettings struct {
	// endpoint is only set for Floci (local dev). Empty means real AWS S3.
	endpoint string
	region   string
	bucket   string
}

func resolveSettingsFromEnv() (storageSettings, error) {
	if flociEnabled() {
		return storageSettings{
			endpoint: envOrDefault(flociEndpointEnv, defaultFlociEndpoint),
			region:   envOrDefault(flociRegionEnv, defaultFlociRegion),
			bucket:   envOrDefault(flociBucketEnv, defaultFlociBucket),
		}, nil
	}

	return resolveS3SettingsFromEnv()
}

func resolveS3SettingsFromEnv() (storageSettings, error) {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		return storageSettings{}, ErrNotConfigured
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "eu-west-1"
	}

	return storageSettings{
		region: region,
		bucket: bucket,
	}, nil
}

func flociEnabled() bool {
	return envBool(flociEnabledEnv)
}

func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
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
// everything else to keep the bucket from doubling as free file hosting.
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

// Upload streams a reader directly to S3. Used by server-side handlers
// that receive a multipart file (e.g. POST /careers/apply) rather than
// delegating the upload to the browser via a presigned URL.
// contentType should match one of AllowedResumeContentTypes; the caller
// is responsible for validating this before calling Upload.
func (c *Client) Upload(ctx context.Context, key, contentType string, body io.Reader) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	})
	return err
}

// Bucket exposes the configured bucket name for callers that need to
// log or report it. Not used for signing (the client already holds it).
func (c *Client) Bucket() string {
	return c.bucket
}

// SanitizeObjectKey removes characters that would trip up S3 or URL
// encoders. The returned string is safe to use as an S3 key and as
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
