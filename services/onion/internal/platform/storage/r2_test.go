package storage

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func resetDefaultStorageForTest() {
	defaultClient = nil
	defaultErr = nil
	defaultOnce = sync.Once{}
}

func TestDefaultUsesFlociWhenEnabled(t *testing.T) {
	t.Setenv("FLOCI_ENABLED", "true")
	t.Setenv("FLOCI_ENDPOINT", "http://localhost:4566")
	t.Setenv("FLOCI_BUCKET", "onion-dev")
	t.Setenv("R2_ACCOUNT_ID", "")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	t.Setenv("R2_BUCKET_NAME", "")
	resetDefaultStorageForTest()
	t.Cleanup(resetDefaultStorageForTest)

	client, err := Default()
	if err != nil {
		t.Fatalf("Default() returned error: %v", err)
	}
	if got, want := client.Bucket(), "onion-dev"; got != want {
		t.Fatalf("Bucket() = %q, want %q", got, want)
	}

	url, err := client.PresignUpload(context.Background(), "resumes/test.pdf", "application/pdf", time.Minute)
	if err != nil {
		t.Fatalf("PresignUpload() returned error: %v", err)
	}
	if !strings.HasPrefix(url, "http://localhost:4566/onion-dev/resumes/test.pdf?") {
		t.Fatalf("PresignUpload() URL = %q, want floci path-style URL", url)
	}
}
