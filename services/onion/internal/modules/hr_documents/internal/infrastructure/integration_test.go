package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_documents/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListDocuments_ForEmployee4101_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListDocuments(ctx, testutil.OwnerID, 4101, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = records
}

func TestListDocuments_AllBelongToEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListDocuments(ctx, testutil.OwnerID, 4101, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	for _, rec := range records {
		if rec.EmployeeID != 4101 {
			t.Errorf("expected employee 4101, got %d", rec.EmployeeID)
		}
		if rec.OwnerID != testutil.OwnerID {
			t.Errorf("record %d has wrong owner: %d", rec.ID, rec.OwnerID)
		}
	}
}

func TestListDocuments_UnknownEmployeeReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListDocuments(ctx, testutil.OwnerID, 999999999, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Errorf("expected empty result for unknown employee, got total=%d records=%d", total, len(records))
	}
}

func TestListDocuments_CountMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListDocuments(ctx, testutil.OwnerID, 4101, 50, 0)
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if int64(len(records)) > total {
		t.Errorf("returned %d records but total reported as %d", len(records), total)
	}
}

func TestGetDocument_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetDocument(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent document, got nil")
	}
	if err != domain.ErrDocumentNotFound {
		t.Errorf("expected ErrDocumentNotFound, got %v", err)
	}
}

func TestCreateAndGetDocument_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	params := domain.CreateDocumentParams{
		EmployeeID:   4101,
		DocumentType: "contract",
		Title:        "Integration Test Doc",
		StorageMode:  "url",
		FileURL:      "https://example.com/doc.pdf",
		UploadedBy:   4101,
	}

	created, err := repo.CreateDocument(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetDocument(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.Title != "Integration Test Doc" {
		t.Errorf("title mismatch: got %q", got.Title)
	}

	// Cleanup
	_ = repo.DeleteDocument(ctx, testutil.OwnerID, created.ID)
}

func TestDeleteDocument_RemovesRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateDocument(ctx, testutil.OwnerID, domain.CreateDocumentParams{
		EmployeeID:   4101,
		DocumentType: "id_card",
		Title:        "To Be Deleted",
		StorageMode:  "url",
		UploadedBy:   4101,
	})
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}

	if err := repo.DeleteDocument(ctx, testutil.OwnerID, created.ID); err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}

	_, err = repo.GetDocument(ctx, testutil.OwnerID, created.ID)
	if err != domain.ErrDocumentNotFound {
		t.Errorf("expected ErrDocumentNotFound after delete, got %v", err)
	}
}

func TestListExpiringDocuments_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListExpiringDocuments(ctx, testutil.OwnerID, 30, 50, 0)
	if err != nil {
		t.Fatalf("ListExpiringDocuments: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = records
}

func TestListExpiringDocuments_AllHaveExpiryDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListExpiringDocuments(ctx, testutil.OwnerID, 365, 50, 0)
	if err != nil {
		t.Fatalf("ListExpiringDocuments: %v", err)
	}
	for _, rec := range records {
		if rec.ExpiryDate == nil {
			t.Errorf("record %d returned from ListExpiringDocuments but has nil ExpiryDate", rec.ID)
		}
	}
}

func TestListDocuments_PaginationOffsetWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, total, err := repo.ListDocuments(ctx, testutil.OwnerID, 4101, 2, 0)
	if err != nil {
		t.Fatalf("ListDocuments page 1: %v", err)
	}
	second, _, err := repo.ListDocuments(ctx, testutil.OwnerID, 4101, 2, 2)
	if err != nil {
		t.Fatalf("ListDocuments page 2: %v", err)
	}
	if total > 2 && len(first) > 0 && len(second) > 0 {
		if first[0].ID == second[0].ID {
			t.Error("pagination offset had no effect")
		}
	}
}

func TestGetDocument_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create under the fixture owner, then try to read under a different owner.
	created, err := repo.CreateDocument(ctx, testutil.OwnerID, domain.CreateDocumentParams{
		EmployeeID:   4101,
		DocumentType: "other",
		Title:        "Owner Isolation Check",
		StorageMode:  "url",
		UploadedBy:   4101,
	})
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	defer func() { _ = repo.DeleteDocument(ctx, testutil.OwnerID, created.ID) }()

	_, err = repo.GetDocument(ctx, 0, created.ID)
	if err != domain.ErrDocumentNotFound {
		t.Errorf("expected ErrDocumentNotFound for wrong owner, got %v", err)
	}
}
