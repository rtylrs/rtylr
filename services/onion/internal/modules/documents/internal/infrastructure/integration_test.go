package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/documents/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListDocuments_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = docs
}

func TestListDocuments_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d for doc id=%d", testutil.OwnerID, d.OwnerID, d.ID)
		}
	}
}

func TestListDocuments_NamesAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.Name == "" {
			t.Errorf("document id=%d has empty Name", d.ID)
		}
	}
}

func TestCreateDocument_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Integration Test Doc",
		Type: "contract",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
	defer func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	}()
}

func TestCreateDocument_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Appears In List Doc",
		Type: "policy",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	}()

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, d := range docs {
		if d.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected document id=%d to appear in List", id)
	}
}

func TestCreateDocument_WithStaffID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name:    "Staff Doc",
		Type:    "id",
		StaffID: 4101,
	})
	if err != nil {
		t.Fatalf("Create with staff: %v", err)
	}
	defer func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	}()
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateDocument_RejectsForeignStaffWithoutInserting(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const name = "Foreign Staff Document Must Not Persist"

	_, err := repo.Create(ctx, testutil.OtherBusinessID, domain.DocumentInput{
		Name:    name,
		StaffID: 4101,
	})
	if !errors.Is(err, domain.ErrDocumentStaffNotFound) {
		t.Fatalf("Create with foreign staff error = %v, want ErrDocumentStaffNotFound", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_documents WHERE business_id = ? AND name = ?`,
		testutil.OtherBusinessID, name,
	).Scan(&count); err != nil {
		t.Fatalf("count rejected document: %v", err)
	}
	if count != 0 {
		t.Fatalf("foreign-staff document persisted %d row(s), want 0", count)
	}
}

func TestCreateDocument_WithFileURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name:    "File URL Doc",
		FileURL: "https://example.com/doc.pdf",
	})
	if err != nil {
		t.Fatalf("Create with file url: %v", err)
	}
	defer func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	}()

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.ID == id {
			if d.FileURL != "https://example.com/doc.pdf" {
				t.Errorf("expected FileURL='https://example.com/doc.pdf', got %q", d.FileURL)
			}
			return
		}
	}
	t.Errorf("document id=%d not found after create", id)
}

func TestDeleteDocument_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "To Delete Doc",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.ID == id {
			t.Errorf("expected document id=%d to be deleted", id)
		}
	}
}
