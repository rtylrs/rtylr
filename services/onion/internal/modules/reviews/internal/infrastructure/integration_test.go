package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	"github.com/voxire/onion/internal/modules/reviews/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const fixtureStaffIDForReviews int64 = 4101 // Alice

func TestListReviews_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: unexpected error: %v", err)
	}
	if reviews == nil {
		t.Fatal("reviews.List: expected non-nil slice, got nil")
	}
}

func TestListReviews_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	reviews, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("reviews.List(ownerID=0): unexpected error: %v", err)
	}
	if len(reviews) != 0 {
		t.Fatalf("reviews.List(ownerID=0): expected 0 rows, got %d", len(reviews))
	}
}

func TestCreateReview_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Period:  "Q2-2026",
		Rating:  4.5,
		Status:  "draft",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("reviews.Create: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("reviews.Create: expected positive ID, got %d", id)
	}
}

func TestCreateReview_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Period:  "Q3-2026",
		Rating:  3.0,
		Status:  "draft",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	var found bool
	for _, r := range reviews {
		if r.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created review id=%d not found in list", id)
	}
}

func TestCreateReview_StaffIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Period:  "Q1-2026",
		Rating:  5.0,
		Status:  "draft",
	}

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	for _, r := range reviews {
		if r.ID == id {
			if r.StaffID != fixtureStaffIDForReviews {
				t.Errorf("StaffID mismatch: got %d, want %d", r.StaffID, fixtureStaffIDForReviews)
			}
			return
		}
	}
	t.Fatalf("created review id=%d not found in list", id)
}

func TestCreateReview_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	for _, r := range reviews {
		if r.ID == id {
			if r.OwnerID != testutil.OwnerID {
				t.Errorf("OwnerID mismatch: got %d, want %d", r.OwnerID, testutil.OwnerID)
			}
			return
		}
	}
	t.Fatalf("created review id=%d not found in list", id)
}

func TestCreateReview_DefaultStatusIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Status left empty -- repo should default to "draft"
	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	for _, r := range reviews {
		if r.ID == id {
			if r.Status != "draft" {
				t.Errorf("expected default status=draft, got %q", r.Status)
			}
			return
		}
	}
	t.Fatalf("created review id=%d not found in list", id)
}

func TestListReviews_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("seed review: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	for _, r := range reviews {
		if r.ID <= 0 {
			t.Errorf("review has non-positive ID: %d", r.ID)
		}
	}
}

func TestCreateReview_WithReviewerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const reviewerID int64 = 4102 // Bob

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID:    fixtureStaffIDForReviews,
		ReviewerID: reviewerID,
		Status:     "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}

	for _, r := range reviews {
		if r.ID == id {
			if r.ReviewerID != reviewerID {
				t.Errorf("ReviewerID mismatch: got %d, want %d", r.ReviewerID, reviewerID)
			}
			return
		}
	}
	t.Fatalf("created review id=%d not found in list", id)
}

func TestListReviews_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("seed review: %v", err)
	}

	other, err := repo.List(ctx, testutil.OwnerID+9999)
	if err != nil {
		t.Fatalf("reviews.List(other owner): %v", err)
	}
	for _, r := range other {
		if r.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}

func TestCreateMultipleReviews_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List before: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
			StaffID: fixtureStaffIDForReviews,
			Status:  "draft",
		})
		if err != nil {
			t.Fatalf("reviews.Create[%d]: %v", i, err)
		}
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List after: %v", err)
	}

	if len(after) <= len(before) {
		t.Fatalf("expected list to grow: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateReview_RatingPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Rating:  4.0,
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.Rating != 4.0 {
				t.Errorf("Rating = %v, want 4.0", r.Rating)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_PeriodPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Period:  "Q4-2026",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.Period != "Q4-2026" {
				t.Errorf("Period = %q, want Q4-2026", r.Period)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_NotesPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Notes:   "Integration test notes",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.Notes != "Integration test notes" {
				t.Errorf("Notes = %q, want %q", r.Notes, "Integration test notes")
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_GoalsRatingPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID:     fixtureStaffIDForReviews,
		GoalsRating: 3.5,
		Status:      "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.GoalsRating != 3.5 {
				t.Errorf("GoalsRating = %v, want 3.5", r.GoalsRating)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_ValuesRatingPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID:      fixtureStaffIDForReviews,
		ValuesRating: 5.0,
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.ValuesRating != 5.0 {
				t.Errorf("ValuesRating = %v, want 5.0", r.ValuesRating)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_SkillsRatingPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID:      fixtureStaffIDForReviews,
		SkillsRating: 2.5,
		Status:       "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.SkillsRating != 2.5 {
				t.Errorf("SkillsRating = %v, want 2.5", r.SkillsRating)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestCreateReview_StatusSubmittedPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "submitted",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.Status != "submitted" {
				t.Errorf("Status = %q, want submitted", r.Status)
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}

func TestListReviews_AllHaveStaffID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.StaffID == 0 {
			t.Errorf("review id=%d has zero StaffID", r.ID)
		}
	}
}

func TestListReviews_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("review id=%d has OwnerID=%d, want %d", r.ID, r.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateReview_CreatedAtIsPopulated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ReviewInput{
		StaffID: fixtureStaffIDForReviews,
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("reviews.Create: %v", err)
	}

	reviews, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("reviews.List: %v", err)
	}
	for _, r := range reviews {
		if r.ID == id {
			if r.CreatedAt.IsZero() {
				t.Error("expected CreatedAt to be non-zero")
			}
			return
		}
	}
	t.Fatalf("review id=%d not found", id)
}
