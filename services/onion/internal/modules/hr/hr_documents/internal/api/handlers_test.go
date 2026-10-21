package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
)

// UploadDocumentHandler.Validate is the only pure validation entry point in
// this package; all other handlers embed the auth check inside Serve and
// therefore require a running service.  These tests cover all documented
// validation rules without touching the database.

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestUploadDocumentHandler_Validate_MissingEmployeeID(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		Title:       "Passport",
		StorageMode: "stored",
	})
	if se == nil {
		t.Fatal("expected validation error, got nil")
	}
	if oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", oneKitStatus(se))
	}
}

func TestUploadDocumentHandler_Validate_MissingTitle(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  1,
		StorageMode: "stored",
	})
	if se == nil {
		t.Fatal("expected validation error for missing title")
	}
	if oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", oneKitStatus(se))
	}
}

func TestUploadDocumentHandler_Validate_MissingStorageMode(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID: 1,
		Title:      "Contract",
	})
	if se == nil {
		t.Fatal("expected validation error for missing storage_mode")
	}
	if oneKitStatus(se) != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", oneKitStatus(se))
	}
}

func TestUploadDocumentHandler_Validate_AllFieldsPresent_ReturnsNil(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  1,
		Title:       "Contract",
		StorageMode: "stored",
		FileURL:     "https://storage.example.com/file.pdf",
	})
	if se != nil {
		t.Errorf("expected no validation error, got %v", se)
	}
}

func TestUploadDocumentHandler_Validate_WhitespaceTitle_Rejected(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  1,
		Title:       "   ",
		StorageMode: "stored",
	})
	if se == nil {
		t.Fatal("expected validation error for whitespace-only title")
	}
}

func TestUploadDocumentHandler_Validate_WhitespaceStorageMode_Rejected(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  1,
		Title:       "Passport",
		StorageMode: "  ",
	})
	if se == nil {
		t.Fatal("expected validation error for whitespace-only storage_mode")
	}
}

func TestUploadDocumentHandler_Validate_LinkedMode_ReturnsNil(t *testing.T) {
	h := &UploadDocumentHandler{}
	se := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  5,
		Title:       "External Link",
		StorageMode: "linked",
		ExternalURL: "https://drive.google.com/doc/123",
	})
	if se != nil {
		t.Errorf("expected nil for valid linked request, got %v", se)
	}
}

func TestUploadDocumentHandler_Validate_ExpiryDate(t *testing.T) {
	expiryDate := "2027-01-01"
	h := &UploadDocumentHandler{}
	if err := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  5,
		Title:       "Passport",
		StorageMode: "stored",
		FileURL:     "https://storage.example.com/passport.pdf",
		ExpiryDate:  &expiryDate,
	}); err != nil {
		t.Fatalf("expected valid expiry date, got %v", err)
	}
}

func TestUploadDocumentHandler_Validate_RejectsInvalidExpiryDate(t *testing.T) {
	expiryDate := "2027-02-30"
	h := &UploadDocumentHandler{}
	err := h.Validate(context.Background(), &UploadDocumentRequest{
		EmployeeID:  5,
		Title:       "Passport",
		StorageMode: "stored",
		FileURL:     "https://storage.example.com/passport.pdf",
		ExpiryDate:  &expiryDate,
	})
	if oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid expiry date, got %v", err)
	}
}
