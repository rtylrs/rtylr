package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/business/internal/domain"
)

// slugUpdateCall records what the repository was asked to write. rows is the
// row count the fake database reports, so a test can reproduce the 0-rows
// outcome the owner predicate produces for another tenant's business.
type slugUpdateCall struct {
	businessID int64
	ownerID    int64
	slug       string
	rows       int64
}

type mockRepo struct {
	profile domain.BusinessProfile
	err     error
	got     int64
	gets    int
	updated slugUpdateCall
	rename  struct {
		businessID int64
		ownerID    int64
		name       string
	}
	setLogo struct {
		businessID int64
		ownerID    int64
		logoURL    string
	}
	deactivate struct {
		businessID int64
		ownerID    int64
	}
	activeOwned int64
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

func (m *mockRepo) UpdateSlug(_ context.Context, businessID, ownerID int64, slug string) (bool, error) {
	m.updated.businessID = businessID
	m.updated.ownerID = ownerID
	m.updated.slug = slug
	return m.updated.rows > 0, m.err
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

func (m *mockRepo) RenameBusiness(_ context.Context, businessID, ownerID int64, name string) (bool, error) {
	m.rename.businessID = businessID
	m.rename.ownerID = ownerID
	m.rename.name = name
	return m.err == nil, m.err
}

func (m *mockRepo) SetBusinessLogo(_ context.Context, businessID, ownerID int64, logoURL string) (bool, error) {
	m.setLogo.businessID = businessID
	m.setLogo.ownerID = ownerID
	m.setLogo.logoURL = logoURL
	return m.err == nil, m.err
}

func (m *mockRepo) DeactivateBusiness(_ context.Context, businessID, ownerID int64) (bool, error) {
	m.deactivate.businessID = businessID
	m.deactivate.ownerID = ownerID
	return m.err == nil, m.err
}

func (m *mockRepo) CountActiveOwnedBusinesses(_ context.Context, _ int64) (int64, error) {
	if m.activeOwned == 0 {
		return 2, m.err
	}
	return m.activeOwned, m.err
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
	businessID := int64(424242)
	ownerID := int64(424243)
	repo := &mockRepo{
		profile: domain.BusinessProfile{SiteName: "Before", Currency: "USD"},
		updated: slugUpdateCall{rows: 1},
	}
	s := &Service{repo: repo}

	first, err := s.GetProfile(ctx, businessID)
	if err != nil {
		t.Fatalf("prime cache: %v", err)
	}
	if first.SiteName != "Before" {
		t.Fatalf("expected primed profile, got %+v", first)
	}

	repo.profile = domain.BusinessProfile{SiteName: "After", Currency: "USD"}

	if err := s.UpdateSlug(ctx, businessID, ownerID, "voxire"); err != nil {
		t.Fatalf("update slug: %v", err)
	}

	after, err := s.GetProfile(ctx, businessID)
	if err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if after.SiteName != "After" {
		t.Fatalf("expected cache invalidation to reload updated profile, got %+v", after)
	}
	if repo.gets != 2 {
		t.Fatalf("expected repository to be hit twice, got %d", repo.gets)
	}
	if repo.updated.businessID != businessID || repo.updated.ownerID != ownerID || repo.updated.slug != "voxire" {
		t.Fatalf("unexpected update args: %+v", repo.updated)
	}
}

// The owner predicate lives in SQL, so a write against another tenant's
// business matches no row. The service must report that instead of returning
// success, which is what used to let a cross-tenant rename look accepted.
func TestUpdateSlug_ForeignBusinessReportsNotOwned(t *testing.T) {
	repo := &mockRepo{} // rows == 0: the owner predicate rejected the write
	s := &Service{repo: repo}

	err := s.UpdateSlug(context.Background(), 992001, 1001, "hijacked")
	if !errors.Is(err, ErrBusinessNotOwned) {
		t.Fatalf("expected ErrBusinessNotOwned, got %v", err)
	}
	if repo.updated.ownerID != 1001 || repo.updated.businessID != 992001 {
		t.Fatalf("write was not scoped to the caller: %+v", repo.updated)
	}
}

func TestUpdateSlug_PropagatesRepositoryError(t *testing.T) {
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{err: want}}

	if err := s.UpdateSlug(context.Background(), 12, 34, "voxire"); !errors.Is(err, want) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestRenameBusiness_TrimsNameAndScopesToOwner(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}

	if err := s.RenameBusiness(context.Background(), 12, 34, "  New name  "); err != nil {
		t.Fatalf("rename business: %v", err)
	}

	if repo.rename.businessID != 12 || repo.rename.ownerID != 34 || repo.rename.name != "New name" {
		t.Fatalf("unexpected rename args: %+v", repo.rename)
	}
}

func TestRenameBusiness_RejectsBlankName(t *testing.T) {
	s := &Service{repo: &mockRepo{}}

	if err := s.RenameBusiness(context.Background(), 12, 34, "   "); !errors.Is(err, ErrBusinessNameEmpty) {
		t.Fatalf("expected blank name error, got %v", err)
	}
}

func TestDeleteBusiness_RejectsOnlyOwnedBusiness(t *testing.T) {
	repo := &mockRepo{activeOwned: 1}
	s := &Service{repo: repo}

	if err := s.DeleteBusiness(context.Background(), 12, 34); !errors.Is(err, ErrCannotDeleteOnlyBusiness) {
		t.Fatalf("expected only-business error, got %v", err)
	}
	if repo.deactivate.businessID != 0 {
		t.Fatal("should not deactivate the only owned business")
	}
}

func TestDeleteBusiness_ScopesToOwner(t *testing.T) {
	repo := &mockRepo{activeOwned: 2}
	s := &Service{repo: repo}

	if err := s.DeleteBusiness(context.Background(), 12, 34); err != nil {
		t.Fatalf("delete business: %v", err)
	}
	if repo.deactivate.businessID != 12 || repo.deactivate.ownerID != 34 {
		t.Fatalf("unexpected deactivate args: %+v", repo.deactivate)
	}
}
