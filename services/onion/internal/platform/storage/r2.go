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

	"github.com/voxire/onion/internal/platform/config"
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
				o.UsePathStyle = settings.forcePathStyle
			})
		}

		defaultClient = &Client{s3: s3.NewFromConfig(cfg, s3Opts...), bucket: settings.bucket}
	})
	return defaultClient, defaultErr
}

type storageSettings struct {
	// endpoint is set for any S3-compatible store (Floci/local dev or an
	// operator-provided endpoint). Empty means real AWS S3.
	endpoint       string
	region         string
	bucket         string
	forcePathStyle bool
}

func resolveSettingsFromEnv() (storageSettings, error) {
	if flociEnabled() {
		endpoint := defaultFlociEndpoint
		region := defaultFlociRegion
		bucket := defaultFlociBucket
		if config.GlobalConfig != nil {
			if config.GlobalConfig.FlociEndpoint != "" {
				endpoint = config.GlobalConfig.FlociEndpoint
			}
			if config.GlobalConfig.FlociRegion != "" {
				region = config.GlobalConfig.FlociRegion
			}
			if config.GlobalConfig.FlociBucket != "" {
				bucket = config.GlobalConfig.FlociBucket
			}
		} else {
			if val := os.Getenv(flociEndpointEnv); val != "" {
				endpoint = val
			}
			if val := os.Getenv(flociRegionEnv); val != "" {
				region = val
			}
			if val := os.Getenv(flociBucketEnv); val != "" {
				bucket = val
			}
		}
		return storageSettings{
			endpoint:       endpoint,
			region:         region,
			bucket:         bucket,
			forcePathStyle: true,
		}, nil
	}

	return resolveS3SettingsFromEnv()
}

func resolveS3SettingsFromEnv() (storageSettings, error) {
	var bucket string
	if config.GlobalConfig != nil {
		bucket = config.GlobalConfig.S3BucketName
	} else {
		bucket = os.Getenv("S3_BUCKET_NAME")
	}
	if bucket == "" {
		return storageSettings{}, ErrNotConfigured
	}

	var region string
	if config.GlobalConfig != nil {
		region = config.GlobalConfig.S3Region
	} else {
		region = os.Getenv("S3_REGION")
	}
	if region == "" {
		region = "eu-west-1"
	}

	var endpoint string
	var forcePathStyle bool
	if config.GlobalConfig != nil {
		endpoint = config.GlobalConfig.S3Endpoint
		forcePathStyle = config.GlobalConfig.S3ForcePathStyle
	} else {
		endpoint = os.Getenv("S3_ENDPOINT")
		forcePathStyle = strings.EqualFold(strings.TrimSpace(os.Getenv("S3_FORCE_PATH_STYLE")), "true")
	}
	// A custom S3-compatible endpoint almost always needs path-style addressing.
	if endpoint != "" {
		forcePathStyle = true
	}

	return storageSettings{
		endpoint:       endpoint,
		region:         region,
		bucket:         bucket,
		forcePathStyle: forcePathStyle,
	}, nil
}

func flociEnabled() bool {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.FlociEnabled
	}
	value := strings.ToLower(strings.TrimSpace(os.Getenv(flociEnabledEnv)))
	return value == "true" || value == "1" || value == "yes"
}

// AllowedResumeContentTypes is the whitelist for the resume upload
// endpoint. PDFs and Word docs cover 99% of candidates; we reject
// everything else to keep the bucket from doubling as free file hosting.
var AllowedResumeContentTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

// MaxResumeBytes is the largest resume accepted when an application is
// submitted. 10 MB is generous — most PDFs are under 1.
const MaxResumeBytes int64 = 10 * 1024 * 1024

// ObjectMetadata is the subset of object properties callers need to verify
// a direct upload before persisting its key.
type ObjectMetadata struct {
	ContentType string
	Size        int64
}

// PresignUpload returns a short-lived PUT URL the browser can upload
// to directly. The key is caller-controlled; the content-type lock
// means a candidate can't swap a PDF for an executable after signing.
func (c *Client) PresignUpload(ctx context.Context, key, contentType string, contentLength int64, expiresIn time.Duration) (string, error) {
	if contentLength <= 0 {
		return "", fmt.Errorf("content length must be positive")
	}
	presign := s3.NewPresignClient(c.s3)
	req, err := presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
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

// Download streams an object out of S3. The caller owns the returned
// reader and must Close it. Size is the object's Content-Length, or 0
// when S3 doesn't report one. Used by the CDN runmode to serve original
// files and cached image transformations.
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("downloading object key=%s: %w", key, err)
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, size, nil
}

// Metadata reads the content type and size of an object without downloading
// it. Unlike Exists, it returns errors so callers can enforce upload
// validation before storing an object key.
func (c *Client) Metadata(ctx context.Context, key string) (ObjectMetadata, error) {
	out, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectMetadata{}, fmt.Errorf("reading object metadata key=%s: %w", key, err)
	}

	metadata := ObjectMetadata{}
	if out.ContentType != nil {
		metadata.ContentType = *out.ContentType
	}
	if out.ContentLength != nil {
		metadata.Size = *out.ContentLength
	}
	return metadata, nil
}

// Exists reports whether an object is present. A HeadObject error is
// treated as "not present" rather than surfaced — callers use this as a
// cheap cache probe, not an authoritative health check.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Delete removes one object from storage. Missing objects are treated as a
// successful delete by S3, which is what CDN cache invalidation wants.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting object key=%s: %w", key, err)
	}
	return nil
}

// DeletePrefix removes every object under prefix. It is intentionally simple:
// CDN purges are rare control-plane actions, so correctness and easy retrying
// matter more than maximum delete throughput.
func (c *Client) DeletePrefix(ctx context.Context, prefix string) (int, error) {
	pager := s3.NewListObjectsV2Paginator(c.s3, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})

	deleted := 0
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return deleted, fmt.Errorf("listing prefix=%s: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			if err := c.Delete(ctx, *obj.Key); err != nil {
				return deleted, err
			}
			deleted++
		}
	}
	return deleted, nil
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
