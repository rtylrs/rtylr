package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyDocumentTitle = errors.New("document title is required")
	ErrEmptyDocumentType  = errors.New("document type is required")
	ErrInvalidStorageMode = errors.New("invalid storage mode")
)

// IsValidDocumentType returns true when the type is recognised.
func IsValidDocumentType(docType string) bool {
	switch docType {
	case DocumentTypeContract, DocumentTypeIDCard, DocumentTypePassport,
		DocumentTypeCertificate, DocumentTypeTaxForm, DocumentTypeOther:
		return true
	}
	return false
}

// IsValidStorageMode returns true when the storage mode is recognised.
func IsValidStorageMode(mode string) bool {
	switch mode {
	case StorageModeURL, StorageModeExternal, StorageModeUpload:
		return true
	}
	return false
}

// ValidateCreateDocumentParams checks that required document creation fields are present.
func ValidateCreateDocumentParams(p CreateDocumentParams) error {
	if strings.TrimSpace(p.Title) == "" {
		return ErrEmptyDocumentTitle
	}
	if strings.TrimSpace(p.DocumentType) == "" {
		return ErrEmptyDocumentType
	}
	if p.StorageMode != "" && !IsValidStorageMode(p.StorageMode) {
		return ErrInvalidStorageMode
	}
	return nil
}
