package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
)

func TestUploadDocument_InvalidStorageMode_RejectsUnknownMode(t *testing.T) {
	s := &Service{}
	_, err := s.UploadDocument(context.Background(), 1, domain.CreateDocumentParams{
		StorageMode: "ftp",
	})
	if !errors.Is(err, domain.ErrStorageModeInvalid) {
		t.Errorf("expected ErrStorageModeInvalid, got %v", err)
	}
}

func TestUploadDocument_StoredMode_RequiresFileURL(t *testing.T) {
	s := &Service{}
	_, err := s.UploadDocument(context.Background(), 1, domain.CreateDocumentParams{
		StorageMode: "stored",
		FileURL:     "",
	})
	if !errors.Is(err, domain.ErrStorageModeInvalid) {
		t.Errorf("expected ErrStorageModeInvalid when stored and no file URL, got %v", err)
	}
}

func TestUploadDocument_LinkedMode_RequiresExternalURL(t *testing.T) {
	s := &Service{}
	_, err := s.UploadDocument(context.Background(), 1, domain.CreateDocumentParams{
		StorageMode: "linked",
		ExternalURL: "",
	})
	if !errors.Is(err, domain.ErrStorageModeInvalid) {
		t.Errorf("expected ErrStorageModeInvalid when linked and no external URL, got %v", err)
	}
}

func TestUploadDocument_EmptyMode_IsRejected(t *testing.T) {
	s := &Service{}
	_, err := s.UploadDocument(context.Background(), 1, domain.CreateDocumentParams{
		StorageMode: "",
	})
	if !errors.Is(err, domain.ErrStorageModeInvalid) {
		t.Errorf("expected ErrStorageModeInvalid for empty storage mode, got %v", err)
	}
}
