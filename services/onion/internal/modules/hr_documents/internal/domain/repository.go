package domain

import "context"

type Repository interface {
	CreateDocument(ctx context.Context, ownerID int64, params CreateDocumentParams) (*EmployeeDocumentRecord, error)
	GetDocument(ctx context.Context, ownerID, id int64) (*EmployeeDocumentRecord, error)
	ListDocuments(ctx context.Context, ownerID, employeeID int64, limit, offset int32) ([]EmployeeDocumentRecord, int64, error)
	ListExpiringDocuments(ctx context.Context, ownerID int64, days, limit, offset int32) ([]EmployeeDocumentRecord, int64, error)
	DeleteDocument(ctx context.Context, ownerID, id int64) error
}
