package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

type mockBannerRepo struct {
	banners     []domain.Banner
	createdID   int64
	createInput domain.BannerInput
	updateInput domain.BannerInput
	updateID    int64
	deleteID    int64
	businessID  int64
	createErr   error
	updateErr   error
	deleteErr   error
	getErr      error
}

func (m *mockBannerRepo) ListBanners(_ context.Context, businessID int64) ([]domain.Banner, error) {
	m.businessID = businessID
	return m.banners, nil
}

func (m *mockBannerRepo) CreateBanner(_ context.Context, businessID int64, input domain.BannerInput) (int64, error) {
	m.businessID = businessID
	m.createInput = input
	return m.createdID, m.createErr
}

func (m *mockBannerRepo) UpdateBanner(_ context.Context, businessID, id int64, input domain.BannerInput) error {
	m.businessID = businessID
	m.updateID = id
	m.updateInput = input
	return m.updateErr
}

func (m *mockBannerRepo) DeleteBanner(_ context.Context, businessID, id int64) error {
	m.businessID = businessID
	m.deleteID = id
	return m.deleteErr
}

func (m *mockBannerRepo) GetBanner(_ context.Context, businessID, id int64) (domain.Banner, error) {
	m.businessID = businessID
	if m.getErr != nil {
		return domain.Banner{}, m.getErr
	}
	return domain.Banner{ID: id, BusinessID: businessID, Title: "Saved"}, nil
}

func (m *mockBannerRepo) List(_ context.Context, businessID int64) ([]*posmenuitemv1.MenuItem, error) {
	m.businessID = businessID
	return nil, nil
}

func (m *mockBannerRepo) Update(_ context.Context, _ int64, _ []*posmenuitemv1.MenuItem) error {
	return nil
}

func TestCreateBannerNormalizesAndReturnsSavedBanner(t *testing.T) {
	repo := &mockBannerRepo{createdID: 55}
	s := &Service{repo: repo}

	got, err := s.CreateBanner(context.Background(), 11, domain.BannerInput{
		Title:    "  Lunch specials  ",
		Body:     "  Soup today  ",
		ImageURL: "  https://example.test/banner.png  ",
		LinkURL:  "  https://example.test/menu  ",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateBanner returned error: %v", err)
	}
	if got.ID != 55 || got.BusinessID != 11 {
		t.Fatalf("saved banner mismatch: %+v", got)
	}
	if repo.createInput.Title != "Lunch specials" || repo.createInput.Body != "Soup today" {
		t.Fatalf("input was not normalized: %+v", repo.createInput)
	}
}

func TestCreateBannerRejectsInvalidInput(t *testing.T) {
	s := &Service{repo: &mockBannerRepo{}}
	_, err := s.CreateBanner(context.Background(), 11, domain.BannerInput{Title: strings.Repeat("x", maxBannerTitleLength+1)})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdateBannerRejectsInvalidDateWindow(t *testing.T) {
	s := &Service{repo: &mockBannerRepo{}}
	startsAt := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	_, err := s.UpdateBanner(context.Background(), 11, 3, domain.BannerInput{
		Title:    "Window",
		StartsAt: startsAt,
		EndsAt:   startsAt.Add(-time.Hour),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDeleteBannerPropagatesRepositoryError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockBannerRepo{deleteErr: want}}
	if err := s.DeleteBanner(context.Background(), 11, 3); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
