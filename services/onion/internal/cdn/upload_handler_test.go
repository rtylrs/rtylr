package cdn

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// multipartUpload builds a multipart/form-data body with one part.
func multipartUpload(t *testing.T, field, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

func TestUpload_MissingFileField(t *testing.T) {
	c := &CDN{maxUpload: 1 << 20}
	body, ct := multipartUpload(t, "notfile", "x.png", []byte("data"))
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing file field, got %d", rec.Code)
	}
}

func TestUpload_EmptyFileRejected(t *testing.T) {
	c := &CDN{maxUpload: 1 << 20}
	body, ct := multipartUpload(t, "file", "x.png", nil)
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty file, got %d", rec.Code)
	}
}
