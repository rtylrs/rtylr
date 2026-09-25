package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

func TestCreateBannerInputDefaultsActiveWhenOmitted(t *testing.T) {
	input, err := createBannerInput(&hubmenuv1.CreateMenuBannerRequest{Title: "Promo"})
	if err != nil {
		t.Fatalf("createBannerInput returned error: %v", err)
	}
	if !input.IsActive {
		t.Fatal("expected omitted is_active to default true")
	}
}

func TestCreateBannerInputPreservesExplicitInactive(t *testing.T) {
	inactive := false
	input, err := createBannerInput(&hubmenuv1.CreateMenuBannerRequest{Title: "Promo", IsActive: &inactive})
	if err != nil {
		t.Fatalf("createBannerInput returned error: %v", err)
	}
	if input.IsActive {
		t.Fatal("expected explicit false is_active to be preserved")
	}
}

func TestCreateBannerInputRejectsInvalidTime(t *testing.T) {
	_, err := createBannerInput(&hubmenuv1.CreateMenuBannerRequest{Title: "Promo", StartsAt: "tomorrow"})
	if err == nil {
		t.Fatal("expected invalid RFC3339 error")
	}
}

func TestToProtoBannerFormatsTimesAsUTC(t *testing.T) {
	banner := toProtoBanner(domain.Banner{
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.FixedZone("EET", 2*60*60)),
	})
	if banner.GetCreatedAt() != "2026-08-13T10:30:00Z" {
		t.Fatalf("unexpected created_at: %q", banner.GetCreatedAt())
	}
}
