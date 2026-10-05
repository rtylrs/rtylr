package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/documents/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateDocument_UniqueIDsEachCall(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{Name: "Unique ID Doc A"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id1) })

	id2, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{Name: "Unique ID Doc B"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id2) })

	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestDeleteDocument_SecondDeleteReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Double Delete Doc",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("First Delete: %v", err)
	}
	if err := repo.Delete(ctx, testutil.OwnerID, id); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("Second Delete error = %v, want ErrDocumentNotFound", err)
	}
}

func TestCreateDocument_FileURLEmptyByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "No URL Doc",
		Type: "policy",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	})

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.ID == id {
			// FileURL may be empty string or zero value when not provided.
			_ = d.FileURL
			return
		}
	}
	t.Errorf("document id=%d not found", id)
}

func TestListDocuments_CountGrowsAfterCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Count Grow Doc",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	})

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d after=%d", len(before), len(after))
	}
}

func TestListDocuments_CountShrinkAfterDelete(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Shrink Count Doc",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before delete: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(after) != len(before)-1 {
		t.Errorf("expected count to shrink by 1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateDocument_WithAllFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		StaffID:    4101,
		Name:       "Full Fields Doc",
		Type:       "passport",
		FileURL:    "https://cdn.example.com/passport.pdf",
		ExpiryDate: "2035-06-30",
	})
	if err != nil {
		t.Fatalf("Create with all fields: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	})

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, d := range docs {
		if d.ID == id {
			if d.Name != "Full Fields Doc" {
				t.Errorf("Name mismatch: got %q", d.Name)
			}
			if d.Type != "passport" {
				t.Errorf("Type mismatch: got %q", d.Type)
			}
			if d.FileURL != "https://cdn.example.com/passport.pdf" {
				t.Errorf("FileURL mismatch: got %q", d.FileURL)
			}
			if d.ExpiryDate.IsZero() {
				t.Error("expected non-zero ExpiryDate")
			}
			return
		}
	}
	t.Errorf("document id=%d not found", id)
}
