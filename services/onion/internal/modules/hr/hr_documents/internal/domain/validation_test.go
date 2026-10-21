package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/domain"
)

func TestIsValidDocumentType_AllValid(t *testing.T) {
	for _, dt := range []string{
		domain.DocumentTypeContract,
		domain.DocumentTypeIDCard,
		domain.DocumentTypePassport,
		domain.DocumentTypeCertificate,
		domain.DocumentTypeTaxForm,
		domain.DocumentTypeOther,
	} {
		if !domain.IsValidDocumentType(dt) {
			t.Errorf("expected %q to be valid", dt)
		}
	}
}

func TestIsValidDocumentType_Invalid(t *testing.T) {
	if domain.IsValidDocumentType("invoice") {
		t.Error("expected 'invoice' to be invalid")
	}
}

func TestIsValidStorageMode_AllValid(t *testing.T) {
	for _, m := range []string{
		domain.StorageModeURL,
		domain.StorageModeExternal,
		domain.StorageModeUpload,
	} {
		if !domain.IsValidStorageMode(m) {
			t.Errorf("expected %q to be valid", m)
		}
	}
}

func TestIsValidStorageMode_Invalid(t *testing.T) {
	if domain.IsValidStorageMode("s3") {
		t.Error("expected 's3' to be invalid")
	}
}

func TestValidateCreateDocumentParams_Valid(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "Employment Contract",
		DocumentType: domain.DocumentTypeContract,
		StorageMode:  domain.StorageModeURL,
		FileURL:      "https://example.com/contract.pdf",
		EmployeeID:   4101,
		UploadedBy:   4101,
	}
	if err := domain.ValidateCreateDocumentParams(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCreateDocumentParams_EmptyTitle(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "",
		DocumentType: domain.DocumentTypeContract,
	}
	if err := domain.ValidateCreateDocumentParams(p); !errors.Is(err, domain.ErrEmptyDocumentTitle) {
		t.Errorf("expected ErrEmptyDocumentTitle, got %v", err)
	}
}

func TestValidateCreateDocumentParams_WhitespaceTitle(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "   ",
		DocumentType: domain.DocumentTypeIDCard,
	}
	if err := domain.ValidateCreateDocumentParams(p); !errors.Is(err, domain.ErrEmptyDocumentTitle) {
		t.Errorf("expected ErrEmptyDocumentTitle for whitespace title, got %v", err)
	}
}

func TestValidateCreateDocumentParams_EmptyDocumentType(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "Some Doc",
		DocumentType: "",
	}
	if err := domain.ValidateCreateDocumentParams(p); !errors.Is(err, domain.ErrEmptyDocumentType) {
		t.Errorf("expected ErrEmptyDocumentType, got %v", err)
	}
}

func TestValidateCreateDocumentParams_InvalidStorageMode(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "Doc",
		DocumentType: domain.DocumentTypeOther,
		StorageMode:  "ftp",
	}
	if err := domain.ValidateCreateDocumentParams(p); !errors.Is(err, domain.ErrInvalidStorageMode) {
		t.Errorf("expected ErrInvalidStorageMode, got %v", err)
	}
}

func TestValidateCreateDocumentParams_EmptyStorageModeAllowed(t *testing.T) {
	p := domain.CreateDocumentParams{
		Title:        "Doc",
		DocumentType: domain.DocumentTypeOther,
		StorageMode:  "",
	}
	if err := domain.ValidateCreateDocumentParams(p); err != nil {
		t.Errorf("empty storage mode should be allowed, got %v", err)
	}
}
