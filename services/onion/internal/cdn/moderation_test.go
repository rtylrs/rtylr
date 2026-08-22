package cdn

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/platform/config"
)

type fakeModerator struct {
	blocked bool
	reason  string
	err     error
}

func (f *fakeModerator) Check(_ context.Context, _ string, _ []byte) (bool, string, error) {
	return f.blocked, f.reason, f.err
}

// pngBytes renders a small solid-color PNG for tests.
func pngBytes(t *testing.T, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestModerateImage_NilModeratorPassesThrough(t *testing.T) {
	c := &CDN{}
	in := bytes.NewReader([]byte("payload"))
	out, err := c.moderateImage(context.Background(), "image/png", 7, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != io.Reader(in) {
		t.Fatal("expected body to pass through untouched")
	}
}

func TestModerateImage_NonImageSkipped(t *testing.T) {
	c := &CDN{moderator: &fakeModerator{blocked: true}}
	in := bytes.NewReader([]byte("%PDF-"))
	if _, err := c.moderateImage(context.Background(), "application/pdf", 5, in); err != nil {
		t.Fatalf("non-image should skip moderation, got %v", err)
	}
}

func TestModerateImage_BlockedReturnsSentinel(t *testing.T) {
	c := &CDN{moderator: &fakeModerator{blocked: true, reason: "sexual"}}
	_, err := c.moderateImage(context.Background(), "image/png", 4, bytes.NewReader([]byte("data")))
	if !errors.Is(err, ErrImageBlocked) {
		t.Fatalf("expected ErrImageBlocked, got %v", err)
	}
}

func TestModerateImage_ErrorFailOpen(t *testing.T) {
	c := &CDN{moderator: &fakeModerator{err: errors.New("api down")}}
	payload := []byte("data")
	out, err := c.moderateImage(context.Background(), "image/png", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("fail-open should allow upload, got %v", err)
	}
	got, readErr := io.ReadAll(out)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("body corrupted after moderation: %q", got)
	}
}

func TestModerateImage_ErrorFailClosed(t *testing.T) {
	c := &CDN{
		moderator:            &fakeModerator{err: errors.New("api down")},
		moderationFailClosed: true,
	}
	_, err := c.moderateImage(context.Background(), "image/png", 4, bytes.NewReader([]byte("data")))
	if !errors.Is(err, ErrImageBlocked) {
		t.Fatalf("fail-closed should reject upload, got %v", err)
	}
}

func TestModerateImage_OversizedSkipsCheck(t *testing.T) {
	c := &CDN{moderator: &fakeModerator{blocked: true}}
	in := bytes.NewReader([]byte("data"))
	out, err := c.moderateImage(context.Background(), "image/png", moderationMaxBytes+1, in)
	if err != nil {
		t.Fatalf("oversized image should skip moderation, got %v", err)
	}
	if out != io.Reader(in) {
		t.Fatal("expected oversized body to pass through unbuffered")
	}
}

func TestUpload_BlockedImageReturns422(t *testing.T) {
	c := &CDN{
		maxUpload: 1 << 20,
		moderator: &fakeModerator{blocked: true, reason: "sexual"},
	}
	body, ct := multipartUpload(t, "file", "x.png", pngBytes(t, color.RGBA{R: 200, G: 150, B: 120, A: 255}))
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for blocked image, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLocalNudeModerator_CleanImagePasses(t *testing.T) {
	m := &localNudeModerator{}
	blocked, _, err := m.Check(context.Background(), "image/png", pngBytes(t, color.RGBA{R: 20, G: 60, B: 200, A: 255}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Fatal("solid blue image should not be flagged")
	}
}

func TestLocalNudeModerator_InvalidImageErrors(t *testing.T) {
	m := &localNudeModerator{}
	if _, _, err := m.Check(context.Background(), "image/png", []byte("not an image")); err == nil {
		t.Fatal("expected decode error for invalid image bytes")
	}
}

func TestLocalNudeModerator_RejectsOversizedDimensionsBeforeDecode(t *testing.T) {
	data := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR" +
		"\xff\xff\xff\xff\x00\x00\x00\x02\x08\x06\x00\x00\x00")
	if _, _, err := (&localNudeModerator{}).Check(context.Background(), "image/png", data); err == nil {
		t.Fatal("expected oversized image dimensions to be rejected")
	}
}

func TestNewModeratorFromConfig(t *testing.T) {
	if m, _ := NewModeratorFromConfig(nil); m != nil {
		t.Fatal("nil config should disable moderation")
	}
	if m, _ := NewModeratorFromConfig(&config.AppConfig{}); m != nil {
		t.Fatal("disabled config should return nil moderator")
	}

	m, failClosed := NewModeratorFromConfig(&config.AppConfig{
		CDNModerationEnabled:    true,
		CDNModerationProvider:   "local",
		CDNModerationFailClosed: true,
	})
	if m == nil {
		t.Fatal("expected local moderator")
	}
	if _, ok := m.(*localNudeModerator); !ok {
		t.Fatalf("expected localNudeModerator, got %T", m)
	}
	if !failClosed {
		t.Fatal("expected fail-closed flag to propagate")
	}

	if m, _ := NewModeratorFromConfig(&config.AppConfig{
		CDNModerationEnabled:  true,
		CDNModerationProvider: "bogus",
	}); m != nil {
		t.Fatal("unknown provider should disable moderation")
	}
}
