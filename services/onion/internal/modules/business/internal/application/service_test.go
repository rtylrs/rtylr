package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
)

type mockRepo struct {
	profile domain.BusinessProfile
	err     error
	got     int64
	gets    int
	updated struct {
		ownerID int64
		slug    string
	}
}

func (m *mockRepo) GetProfile(_ context.Context, ownerID int64) (domain.BusinessProfile, error) {
	m.got = ownerID
	m.gets++
	return m.profile, m.err
}

func (m *mockRepo) GetPublicBusiness(_ context.Context, businessID int64) (domain.PublicBusiness, error) {
	return domain.PublicBusiness{ID: businessID}, m.err
}

func (m *mockRepo) LookupBySlug(_ context.Context, _ string) (domain.SlugLookupResult, error) {
	return domain.SlugLookupResult{}, m.err
}

func (m *mockRepo) UpdateSlug(_ context.Context, ownerID int64, slug string) error {
	m.updated.ownerID = ownerID
	m.updated.slug = slug
	return m.err
}

func (m *mockRepo) ListUserBusinesses(_ context.Context, _ int64) ([]domain.OwnedBusiness, error) {
	return nil, m.err
}

func (m *mockRepo) CreateSubBusiness(_ context.Context, _ string, _, _ int64) (int64, error) {
	return 0, m.err
}

func (m *mockRepo) CheckSubBusinessOwnership(_ context.Context, _, _ int64) (bool, error) {
	return false, m.err
}

func TestGetProfile_Delegates(t *testing.T) {
	want := domain.BusinessProfile{SiteName: "Acme", Currency: "USD"}
	m := &mockRepo{profile: want}
	s := &Service{repo: m}

	got, err := s.GetProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SiteName != "Acme" || got.Currency != "USD" {
		t.Errorf("got %+v", got)
	}
	if m.got != 7 {
		t.Errorf("ownerID didn't pass through, got %d", m.got)
	}
}

func TestGetProfile_PropagatesError(t *testing.T) {
	want := errors.New("missing")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetProfile(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateSlug_InvalidatesLocalCachedProfile(t *testing.T) {
	ctx := context.Background()
	ownerID := int64(424242)
	repo := &mockRepo{profile: domain.BusinessProfile{SiteName: "Before", Currency: "USD"}}
	s := &Service{repo: repo}

	first, err := s.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("prime cache: %v", err)
	}
	if first.SiteName != "Before" {
		t.Fatalf("expected primed profile, got %+v", first)
	}

	repo.profile = domain.BusinessProfile{SiteName: "After", Currency: "USD"}

	if err := s.UpdateSlug(ctx, ownerID, "voxire"); err != nil {
		t.Fatalf("update slug: %v", err)
	}

	after, err := s.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if after.SiteName != "After" {
		t.Fatalf("expected cache invalidation to reload updated profile, got %+v", after)
	}
	if repo.gets != 2 {
		t.Fatalf("expected repository to be hit twice, got %d", repo.gets)
	}
	if repo.updated.ownerID != ownerID || repo.updated.slug != "voxire" {
		t.Fatalf("unexpected update args: %+v", repo.updated)
	}
}
