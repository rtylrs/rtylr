package domain

import "time"

type EmployeeDocumentRecord struct {
	ID           int64      `json:"id"`
	OwnerID      int64      `json:"owner_id"`
	BusinessID   int64      `json:"business_id"`
	EmployeeID   int64      `json:"employee_id"`
	DocumentType string     `json:"document_type"`
	Title        string     `json:"title"`
	StorageMode  string     `json:"storage_mode"`
	FileURL      string     `json:"file_url"`
	ExternalURL  string     `json:"external_url"`
	ExpiryDate   *time.Time `json:"expiry_date"`
	UploadedBy   int64      `json:"uploaded_by"`
	EmployeeName string     `json:"employee_name,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateDocumentParams struct {
	EmployeeID   int64      `json:"employee_id"`
	DocumentType string     `json:"document_type"`
	Title        string     `json:"title"`
	StorageMode  string     `json:"storage_mode"`
	FileURL      string     `json:"file_url"`
	ExternalURL  string     `json:"external_url"`
	ExpiryDate   *time.Time `json:"expiry_date"`
	UploadedBy   int64      `json:"uploaded_by"`
}

type ListDocumentsParams struct {
	Limit  int32
	Offset int32
}
