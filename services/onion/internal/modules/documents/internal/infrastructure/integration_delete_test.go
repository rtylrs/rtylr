package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	"github.com/voxire/onion/internal/modules/documents/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestDeleteDocument_WrongOwnerReturnsNotFoundAndPreservesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Isolation Doc",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_ = repo.Delete(ctx, testutil.OwnerID, id)
	}()

	if err := repo.Delete(ctx, 9999999, id); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("Delete with wrong owner error = %v, want ErrDocumentNotFound", err)
	}

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
		t.Errorf("document id=%d was incorrectly deleted by wrong owner", id)
	}
}

func TestListDocuments_WithExpiryDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name:       "Expiring Doc",
		ExpiryDate: "2030-12-31",
	})
	if err != nil {
		t.Fatalf("Create with expiry: %v", err)
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
			if d.ExpiryDate.IsZero() {
				t.Error("expected non-zero ExpiryDate")
			}
			return
		}
	}
	t.Errorf("document id=%d not found", id)
}

func TestCreateDocument_TypePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: "Type Test Doc",
		Type: "invoice",
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
			if d.Type != "invoice" {
				t.Errorf("expected Type='invoice', got %q", d.Type)
			}
			return
		}
	}
	t.Errorf("document id=%d not found after create", id)
}

func TestCreateDocument_StaffIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name:    "Staff Persist Doc",
		Type:    "id",
		StaffID: 4101,
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
			if d.StaffID != 4101 {
				t.Errorf("expected StaffID=4101, got %d", d.StaffID)
			}
			return
		}
	}
	t.Errorf("document id=%d not found after create", id)
}

func TestCreateDocument_NamePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantName = "Name Persisted Doc"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
		Name: wantName,
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
			if d.Name != wantName {
				t.Errorf("expected Name=%q, got %q", wantName, d.Name)
			}
			return
		}
	}
	t.Errorf("document id=%d not found", id)
}

func TestDeleteDocument_NonExistentReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 9999999); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("Delete non-existent error = %v, want ErrDocumentNotFound", err)
	}
}

func TestListDocuments_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	docs, err := repo.List(ctx, 9999999)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, d := range docs {
		if d.OwnerID == testutil.OwnerID {
			t.Errorf("isolation violated: doc id=%d belongs to fixture owner", d.ID)
		}
	}
}

func TestCreateMultipleDocuments_AllAppearInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var created []int64
	for i := 0; i < 3; i++ {
		id, err := repo.Create(ctx, testutil.OwnerID, domain.DocumentInput{
			Name: fmt.Sprintf("Bulk Doc %d", i),
			Type: "contract",
		})
		if err != nil {
			t.Fatalf("Create i=%d: %v", i, err)
		}
		created = append(created, id)
	}
	t.Cleanup(func() {
		for _, id := range created {
			_ = repo.Delete(ctx, testutil.OwnerID, id)
		}
	})

	docs, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := make(map[int64]bool)
	for _, d := range docs {
		found[d.ID] = true
	}
	for _, id := range created {
		if !found[id] {
			t.Errorf("expected document id=%d to appear in list", id)
		}
	}
}
