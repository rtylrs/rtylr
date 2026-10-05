package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateSegment_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "High Value Customers",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateSegment_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Stored Name Segment",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	seg, err := repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSegmentByID: %v", err)
	}
	if seg.Name != "Stored Name Segment" {
		t.Errorf("Name = %q, want %q", seg.Name, "Stored Name Segment")
	}
}

func TestGetSegmentByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	_, err := repo.GetSegmentByID(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrSegmentNotFound) {
		t.Errorf("expected ErrSegmentNotFound, got %v", err)
	}
}

func TestGetSegmentByID_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwner Segment",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	_, err = repo.GetSegmentByID(ctx, 999999, id)
	if err == nil {
		t.Fatal("expected error with wrong ownerID, got nil")
	}
}

func TestUpdateSegment_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Before Name Seg",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	if err := repo.UpdateSegment(ctx, domain.SegmentUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "After Name Seg",
		Rules:      []byte(`{}`),
	}); err != nil {
		t.Fatalf("UpdateSegment: %v", err)
	}

	seg, err := repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSegmentByID: %v", err)
	}
	if seg.Name != "After Name Seg" {
		t.Errorf("Name = %q, want %q", seg.Name, "After Name Seg")
	}
}

func TestUpdateSegment_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwner Update Seg",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	err = repo.UpdateSegment(ctx, domain.SegmentUpdateInput{
		ID:      id,
		OwnerID: 999999,
		Name:    "Sneaky Update",
		Rules:   []byte(`{}`),
	})
	if !errors.Is(err, domain.ErrSegmentNotFound) {
		t.Errorf("expected ErrSegmentNotFound, got %v", err)
	}
}

func TestDeleteSegment_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Delete Me Seg",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	if err := repo.DeleteSegment(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteSegment: %v", err)
	}

	_, err = repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestListSegments_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "List Check Seg",
		Rules:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	segments, err := repo.ListSegments(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}

	found := false
	for _, s := range segments {
		if s.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("segment id=%d not found in list", id)
	}
}

func TestListSegments_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	segments, err := repo.ListSegments(ctx, 999999)
	if err != nil {
		t.Fatalf("ListSegments wrong owner: %v", err)
	}
	if len(segments) != 0 {
		t.Fatalf("expected 0 segments for unknown owner, got %d", len(segments))
	}
}
