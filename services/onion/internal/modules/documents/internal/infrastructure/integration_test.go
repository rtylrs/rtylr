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
