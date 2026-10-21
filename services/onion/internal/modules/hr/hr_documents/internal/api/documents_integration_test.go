package api

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestDocumentHandlers_LiveListShapesAndExpiryPersistence(t *testing.T) {
	db := testutil.OpenDB(t)
	service := application.NewService(infrastructure.NewRepository(db))
	ctx := documentAuthContext("hr:employees:view", "hr:employees:manage")
	expiryDate := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02")

	expiring, err := (&UploadDocumentHandler{Service: service}).Serve(ctx, &UploadDocumentRequest{
		EmployeeID:   4101,
		DocumentType: "passport",
		Title:        "Live Expiry Contract",
		StorageMode:  "stored",
		FileURL:      "https://example.com/live-expiry.pdf",
		ExpiryDate:   &expiryDate,
	})
	if err != nil {
		t.Fatalf("upload expiring document: %v", err)
	}
	t.Cleanup(func() { _ = service.DeleteDocument(context.Background(), testutil.BusinessID, expiring.Document.Id) })
	if expiring.Document.ExpiryDate == nil || *expiring.Document.ExpiryDate != expiryDate {
		t.Fatalf("upload response expiry = %v, want %s", expiring.Document.ExpiryDate, expiryDate)
	}

	second, err := (&UploadDocumentHandler{Service: service}).Serve(ctx, &UploadDocumentRequest{
		EmployeeID:   4102,
		DocumentType: "contract",
		Title:        "Live All Employees Contract",
		StorageMode:  "stored",
		FileURL:      "https://example.com/live-all.pdf",
	})
	if err != nil {
		t.Fatalf("upload second employee document: %v", err)
	}
	t.Cleanup(func() { _ = service.DeleteDocument(context.Background(), testutil.BusinessID, second.Document.Id) })

	employeeResponse, err := (&ListDocumentsHandler{Service: service}).Serve(ctx, &ListDocumentsRequest{EmployeeID: 4101})
	if err != nil {
		t.Fatalf("employee-scoped list: %v", err)
	}
	page, ok := employeeResponse.(ListDocumentsResponse)
	if !ok {
		t.Fatalf("employee-scoped response type = %T, want ListDocumentsResponse", employeeResponse)
	}
	if page.Total < 1 || len(page.Documents) < 1 {
		t.Fatalf("employee-scoped page is empty: %+v", page)
	}

	allResponse, err := (&ListDocumentsHandler{Service: service}).Serve(ctx, &ListDocumentsRequest{})
	if err != nil {
		t.Fatalf("all-employee list: %v", err)
	}
	legacy, ok := allResponse.([]LegacyDocumentResponse)
	if !ok {
		t.Fatalf("all-employee response type = %T, want []LegacyDocumentResponse", allResponse)
	}
	employees := map[int64]string{}
	for _, document := range legacy {
		employees[document.StaffID] = document.StaffName
	}
	if employees[4101] != "Alice Worker" || employees[4102] != "Bob Hourly" {
		t.Fatalf("all-employee response omitted an employee: %+v", employees)
	}

	expiringResponse, err := (&ListExpiringDocumentsHandler{Service: service}).Serve(ctx, &ListExpiringDocumentsRequest{Days: 30, Limit: 100})
	if err != nil {
		t.Fatalf("expiring list: %v", err)
	}
	expiringLegacy, ok := expiringResponse.([]LegacyDocumentResponse)
	if !ok {
		t.Fatalf("expiring response type = %T, want []LegacyDocumentResponse", expiringResponse)
	}
	found := false
	for _, document := range expiringLegacy {
		if document.ID == expiring.Document.Id {
			found = true
			if document.ExpiresAt != expiryDate {
				t.Fatalf("legacy expiry = %q, want %q", document.ExpiresAt, expiryDate)
			}
		}
	}
	if !found {
		t.Fatalf("persisted expiring document %d was not returned by the live expiring list", expiring.Document.Id)
	}
}
