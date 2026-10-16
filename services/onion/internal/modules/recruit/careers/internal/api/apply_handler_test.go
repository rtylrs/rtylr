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

// job_id=0 rows join no opening in the recruiter pipeline, so a missing or
// malformed job_id must fail the request instead of being stored as an orphan.
func TestApplyHandlerRejectsMissingJobID(t *testing.T) {
	body, contentType := newCareerMultipartBodyWithFields(t, 0, map[string]string{"job_id": ""})
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestApplyHandlerRejectsMalformedJobID(t *testing.T) {
	body, contentType := newCareerMultipartBodyWithFields(t, 0, map[string]string{"job_id": "abc"})
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestApplyHandlerRejectsInvalidEmail(t *testing.T) {
	body, contentType := newCareerMultipartBodyWithFields(t, 0, map[string]string{"email": "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// A CV that is not a PDF/Word document used to be dropped silently, so the
// applicant believed it had been attached.
func TestApplyHandlerRejectsUnsupportedCVType(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range map[string]string{"name": "Ada", "email": "ada@example.com", "job_id": "4242"} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField: %v", err)
		}
	}
	part, err := writer.CreateFormFile("cv", "photo.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("png")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/careers/apply", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	ApplyHandler(context.Background(), nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func newCareerMultipartBody(t *testing.T, cvSize int) (*bytes.Buffer, string) {
	t.Helper()
	return newCareerMultipartBodyWithFields(t, cvSize, nil)
}

func newCareerMultipartBodyWithFields(t *testing.T, cvSize int, extraFields map[string]string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fields := map[string]string{
		"name":   "Ada Lovelace",
		"email":  "ada@example.com",
		"job_id": "4242",
	}
	for key, value := range extraFields {
		fields[key] = value
	}
	for key, value := range fields {
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
