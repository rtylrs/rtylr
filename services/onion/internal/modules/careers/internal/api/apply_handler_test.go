package api

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplyHandlerRejectsOversizedMultipartBody(t *testing.T) {
	body, contentType := newCareerMultipartBody(t, maxApplicationMultipartBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestApplyHandlerRejectsOversizedCV(t *testing.T) {
	body, contentType := newCareerMultipartBody(t, maxResumeUploadBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func newCareerMultipartBody(t *testing.T, cvSize int) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range map[string]string{
		"name":  "Ada Lovelace",
		"email": "ada@example.com",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s): %v", key, err)
		}
	}

	part, err := writer.CreateFormFile("cv", "resume.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(bytes.Repeat([]byte("x"), cvSize)); err != nil {
		t.Fatalf("write cv: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	return body, writer.FormDataContentType()
}
