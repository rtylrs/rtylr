package keys

import "testing"

func TestUserMeForBusinessScopesByUserAndBusiness(t *testing.T) {
	got := UserMeForBusiness(42, 7)
	want := "cache:user:42:business:7:me"
	if got != want {
		t.Fatalf("UserMeForBusiness = %q, want %q", got, want)
	}
}

func TestBusinessProfileDoesNotCollideWithConfig(t *testing.T) {
	configKey := BusinessConfig(7)
	profileKey := BusinessProfile(7)

	if profileKey == configKey {
		t.Fatalf("BusinessProfile key collides with BusinessConfig: %q", profileKey)
	}
	if profileKey != "cache:business:7:profile" {
		t.Fatalf("BusinessProfile = %q", profileKey)
	}
}

func TestUserBusinessesIsUserScoped(t *testing.T) {
	got := UserBusinesses(42)
	want := "cache:user:42:businesses"
	if got != want {
		t.Fatalf("UserBusinesses = %q, want %q", got, want)
	}
}

func TestUserBusinessMetadataKeysAreUserScoped(t *testing.T) {
	if got, want := UserBusinessID(42), "cache:user:42:business_id"; got != want {
		t.Fatalf("UserBusinessID = %q, want %q", got, want)
	}
	if got, want := UserBusinessRole(42, 7), "cache:user:42:business:7:role"; got != want {
		t.Fatalf("UserBusinessRole = %q, want %q", got, want)
	}
	if got, want := UserAllowedApps(42, 7), "cache:user:42:business:7:apps"; got != want {
		t.Fatalf("UserAllowedApps = %q, want %q", got, want)
	}
}

func TestBusinessEnabledAddonsIsBusinessScoped(t *testing.T) {
	got := BusinessEnabledAddons(7)
	want := "cache:business:7:enabled_addons"
	if got != want {
		t.Fatalf("BusinessEnabledAddons = %q, want %q", got, want)
	}
}
