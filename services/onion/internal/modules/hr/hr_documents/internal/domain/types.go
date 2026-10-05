package domain

import "time"

type EmployeeDocumentRecord struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	EmployeeID   int64
	DocumentType string
	Title        string
	StorageMode  string
	FileURL      string
	ExternalURL  string
	ExpiryDate   *time.Time
	UploadedBy   int64
	EmployeeName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateDocumentParams struct {
	EmployeeID   int64
	DocumentType string
	Title        string
	StorageMode  string
	FileURL      string
	ExternalURL  string
	ExpiryDate   *time.Time
	UploadedBy   int64
}

type ListDocumentsParams struct {
	Limit  int32
	Offset int32
}
