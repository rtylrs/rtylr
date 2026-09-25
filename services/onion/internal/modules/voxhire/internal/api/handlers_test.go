package api

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContactHandler_Validate_RequiresCompanyName(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "",
		YourName:    "Alice",
		WorkEmail:   "alice@acme.com",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when company_name is empty")
	}
}

func TestContactHandler_Validate_RequiresYourName(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "",
		WorkEmail:   "alice@acme.com",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when your_name is empty")
	}
}

func TestContactHandler_Validate_RequiresWorkEmail(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "Alice",
		WorkEmail:   "",
		Role:        "CTO",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when work_email is empty")
	}
}

func TestContactHandler_Validate_RequiresRole(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme",
		YourName:    "Alice",
		WorkEmail:   "alice@acme.com",
		Role:        "",
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when role is empty")
	}
}

func TestContactHandler_Validate_HappyPath(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "Acme Corp",
		YourName:    "Alice Smith",
		WorkEmail:   "alice@acme.com",
		Role:        "Engineering Manager",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContactHandler_Validate_AllEmpty(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestContactHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &ContactHandler{}
	req := &ContactRequest{
		CompanyName: "StartupCo",
		YourName:    "Bob",
		WorkEmail:   "bob@startup.io",
		Role:        "CEO",
		Headcount:   "10-50",
		Timeline:    "Q3 2026",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestApplicationsHandlerRejectsOversizedMultipartBody(t *testing.T) {
	body, contentType := newVoxhireMultipartBody(t, maxApplicationMultipartBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/voxhire/applications", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplicationsHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestApplicationsHandlerRejectsOversizedCV(t *testing.T) {
	body, contentType := newVoxhireMultipartBody(t, maxResumeUploadBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/voxhire/applications", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplicationsHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestApplicationsHandlerRequiresCV(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range map[string]string{
		"fullName": "Ada Lovelace",
		"email":    "ada@example.com",
		"consent":  "true",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s): %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/voxhire/applications", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	ApplicationsHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestApplicationsHandlerRejectsUnsupportedCV(t *testing.T) {
	body, contentType := newVoxhireMultipartBodyWithFile(t, 10, "notes.txt")
	req := httptest.NewRequest(http.MethodPost, "/voxhire/applications", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	ApplicationsHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func newVoxhireMultipartBody(t *testing.T, cvSize int) (*bytes.Buffer, string) {
	t.Helper()
	return newVoxhireMultipartBodyWithFile(t, cvSize, "resume.pdf")
}

func newVoxhireMultipartBodyWithFile(t *testing.T, cvSize int, filename string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range map[string]string{
		"fullName": "Ada Lovelace",
		"email":    "ada@example.com",
		"consent":  "true",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s): %v", key, err)
		}
	}

	part, err := writer.CreateFormFile("cv", filename)
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
