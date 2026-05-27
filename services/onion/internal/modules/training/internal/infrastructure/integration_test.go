package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/training/internal/domain"
	"github.com/voxire/onion/internal/modules/training/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTraining_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestTraining_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, c.OwnerID)
		}
	}
}

func TestTraining_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:          "Safety Training",
		Description:   "Workplace safety procedures",
		DurationHours: 4.0,
		IsRequired:    true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestTraining_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:          "List Test Course",
		DurationHours: 2.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, c := range items {
		if c.ID == id {
			found = true
			if c.Name != "List Test Course" {
				t.Errorf("expected name 'List Test Course', got %s", c.Name)
			}
		}
	}
	if !found {
		t.Errorf("created course id=%d not found in list", id)
	}
}

func TestTraining_Create_RequiredCourse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:       "Mandatory Compliance",
		IsRequired: true,
	})
	if err != nil {
		t.Fatalf("Create required course: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == id && !c.IsRequired {
			t.Error("expected IsRequired=true")
		}
	}
}

func TestTraining_Create_OptionalCourse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:       "Optional Leadership",
		IsRequired: false,
	})
	if err != nil {
		t.Fatalf("Create optional course: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, c := range items {
		if c.ID == id && c.IsRequired {
			t.Error("expected IsRequired=false")
		}
	}
}

func TestTraining_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List unknown owner: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestTraining_Create_WithDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:          "Described Course",
		Description:   "A detailed description of this training course.",
		DurationHours: 8.0,
		IsRequired:    false,
	})
	if err != nil {
		t.Fatalf("Create with description: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTraining_Create_ZeroDuration(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name:          "No Duration",
		DurationHours: 0.0,
	})
	if err != nil {
		t.Fatalf("Create zero duration: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTraining_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	courses := []domain.TrainingCourseInput{
		{Name: "Course A", DurationHours: 1.0},
		{Name: "Course B", DurationHours: 2.0},
		{Name: "Course C", DurationHours: 3.0},
	}
	for _, c := range courses {
		id, err := repo.Create(context.Background(), testutil.OwnerID, c)
		if err != nil {
			t.Fatalf("Create %s: %v", c.Name, err)
		}
		if id <= 0 {
			t.Errorf("course %s: expected positive ID", c.Name)
		}
	}
}

func TestTraining_Create_UniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Unique A"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Unique B"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}

func TestTraining_List_NamesAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Named Course"})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.Name == "" {
			t.Errorf("course id=%d has empty Name", c.ID)
		}
	}
}

func TestTraining_Create_DurationStoredCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const wantDuration = 6.5
	id, err := repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{
		Name: "Duration Test Course", DurationHours: wantDuration,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			if c.DurationHours != wantDuration {
				t.Errorf("expected DurationHours=%.1f, got %.1f", wantDuration, c.DurationHours)
			}
			return
		}
	}
	t.Errorf("created course id=%d not found", id)
}

func TestTraining_List_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID <= 0 {
			t.Errorf("expected positive ID, got %d", c.ID)
		}
	}
}

func TestTraining_Create_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Owner A Course"})

	otherItems, err := repo.List(ctx, 8888888)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, c := range otherItems {
		if c.OwnerID == testutil.OwnerID {
			t.Errorf("owner isolation violated: found course belonging to OwnerID=%d", c.OwnerID)
		}
	}
}

func TestTraining_Create_LongDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	longDesc := "This is a comprehensive training course covering all aspects of workplace safety, " +
		"including fire drills, evacuation procedures, first aid fundamentals, and OSHA compliance guidelines."
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name: "Long Desc Course", Description: longDesc, DurationHours: 8.0,
	})
	if err != nil {
		t.Fatalf("Create long desc: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTraining_Create_HalfHourDuration(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name: "Quick Onboard", DurationHours: 0.5,
	})
	if err != nil {
		t.Fatalf("Create half-hour: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTraining_Create_LongDuration(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.TrainingCourseInput{
		Name: "Intensive Bootcamp", DurationHours: 80.0, IsRequired: true,
	})
	if err != nil {
		t.Fatalf("Create long duration: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTraining_List_OwnerIDFieldSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Owner Field Test"})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d", testutil.OwnerID, c.OwnerID)
		}
	}
}

func TestTraining_Create_Idempotent_SameNameAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Duplicate Name"})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	id2, err := repo.Create(ctx, testutil.OwnerID, domain.TrainingCourseInput{Name: "Duplicate Name"})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs even for same name, both got %d", id1)
	}
}

func TestTraining_List_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	b, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(a) != len(b) {
		t.Errorf("idempotency: got %d then %d", len(a), len(b))
	}
}
