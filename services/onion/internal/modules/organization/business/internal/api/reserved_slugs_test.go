package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
	businessdomain "github.com/voxire/onion/internal/modules/organization/business/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/business/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

type mockUserService struct {
	existsNames map[string]bool
}

func (m *mockUserService) BusinessNameExists(_ context.Context, name string) (bool, error) {
	return m.existsNames[name], nil
}

func (m *mockUserService) OnboardCreateBusiness(_ context.Context, _ int64, _ string) (string, error) {
	return "token-123", nil
}

func TestSetSlugHandler_RejectsReservedSlugs(t *testing.T) {
	db, _ := newStubDB(t, 1)
	handler := &SetSlugHandler{Service: application.NewService(context.Background(), moduleinfra.NewRepository(db))}

	reservedCases := []string{
		"menu", "menus", "automations", "portal", "pos", "settings",
		"MENU", "  orders  ", "Public", "ADMIN",
	}

	for _, slug := range reservedCases {
		t.Run(slug, func(t *testing.T) {
			_, err := handler.Serve(deviceCtx(callerUserID, callerBusinessID), &SetSlugRequest{Slug: slug})
			if err == nil {
				t.Fatalf("expected error for reserved slug %q, got nil", slug)
			}
			var apiErr *onekiterrors.Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected onekiterrors.Error, got %T: %v", err, err)
			}
			if apiErr.HTTPStatusCode() != http.StatusConflict {
				t.Fatalf("expected 409 Conflict, got %d", apiErr.HTTPStatusCode())
			}
			if apiErr.Error() != "This slug is reserved." {
				t.Fatalf("expected 'This slug is reserved.', got %q", apiErr.Error())
			}
		})
	}
}

func TestBusinessAPI_SetSlug_RejectsReservedSlugs(t *testing.T) {
	db, _ := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	reservedCases := []string{
		"menu", "menus", "help", "chat", "caret", "auth", "login", "signup",
	}

	for _, slug := range reservedCases {
		t.Run(slug, func(t *testing.T) {
			_, err := api.SetSlug(deviceCtx(callerUserID, callerBusinessID), &hubbusinessv1.SetSlugRequest{Slug: slug})
			if err == nil {
				t.Fatalf("expected error for reserved slug %q, got nil", slug)
			}
			var apiErr *onekiterrors.Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected onekiterrors.Error, got %T: %v", err, err)
			}
			if apiErr.HTTPStatusCode() != http.StatusConflict {
				t.Fatalf("expected 409 Conflict, got %d", apiErr.HTTPStatusCode())
			}
			if apiErr.Error() != "This slug is reserved." {
				t.Fatalf("expected 'This slug is reserved.', got %q", apiErr.Error())
			}
		})
	}
}

func TestBusinessAPI_SetSlug_AllowsNormalAndSimilarSlugs(t *testing.T) {
	db, _ := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	allowedCases := []struct {
		input    string
		expected string
	}{
		{input: "my-menu", expected: "my-menu"},
		{input: "menu-bar", expected: "menu-bar"},
		{input: "the-menus", expected: "the-menus"},
		{input: "acme-coffee", expected: "acme-coffee"},
	}

	for _, tc := range allowedCases {
		t.Run(tc.input, func(t *testing.T) {
			resp, err := api.SetSlug(deviceCtx(callerUserID, callerBusinessID), &hubbusinessv1.SetSlugRequest{Slug: tc.input})
			if err != nil {
				t.Fatalf("unexpected error for slug %q: %v", tc.input, err)
			}
			if resp.GetSlug() != tc.expected {
				t.Fatalf("expected slug %q, got %q", tc.expected, resp.GetSlug())
			}
		})
	}
}

func TestBusinessAPI_CheckBusinessNameExists_ReservedSlugs(t *testing.T) {
	db, _ := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))
	api.WithUserService(&mockUserService{
		existsNames: map[string]bool{
			"existing-business": true,
		},
	})

	tests := []struct {
		name       string
		business   string
		wantExists bool
	}{
		{name: "reserved slug menu", business: "menu", wantExists: true},
		{name: "reserved slug menus", business: "menus", wantExists: true},
		{name: "reserved slug portal", business: "portal", wantExists: true},
		{name: "reserved slug case variant MENU", business: "MENU", wantExists: true},
		{name: "reserved slug whitespace variant   pos  ", business: "  pos  ", wantExists: true},
		{name: "non-reserved available my-menu", business: "my-menu", wantExists: false},
		{name: "non-reserved available unique-name", business: "Unique Bistro", wantExists: false},
		{name: "database existing business", business: "existing-business", wantExists: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := api.CheckBusinessNameExists(context.Background(), &hubbusinessv1.CheckBusinessNameExistsRequest{
				Name: tc.business,
			})
			if err != nil {
				t.Fatalf("CheckBusinessNameExists: %v", err)
			}
			if resp.GetExists() != tc.wantExists {
				t.Errorf("CheckBusinessNameExists(%q) = %v, want %v", tc.business, resp.GetExists(), tc.wantExists)
			}
		})
	}
}

func TestEveryReservedSlugRejectedInSetSlug(t *testing.T) {
	db, _ := newStubDB(t, 1)
	api := NewBusinessAPI(application.NewService(context.Background(), moduleinfra.NewRepository(db)))

	for _, slug := range businessdomain.ReservedBusinessSlugsList() {
		t.Run(slug, func(t *testing.T) {
			_, err := api.SetSlug(deviceCtx(callerUserID, callerBusinessID), &hubbusinessv1.SetSlugRequest{Slug: slug})
			if err == nil {
				t.Fatalf("expected SetSlug to reject reserved slug %q", slug)
			}
		})
	}
}
