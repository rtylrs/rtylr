package domain

import "context"

type Repository interface {
	CreateDocument(ctx context.Context, businessID int64, params CreateDocumentParams) (*EmployeeDocumentRecord, error)
	GetDocument(ctx context.Context, businessID, id int64) (*EmployeeDocumentRecord, error)
	ListDocuments(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]EmployeeDocumentRecord, int64, error)
	ListExpiringDocuments(ctx context.Context, businessID int64, days, limit, offset int32) ([]EmployeeDocumentRecord, int64, error)
	DeleteDocument(ctx context.Context, businessID, id int64) error
}
