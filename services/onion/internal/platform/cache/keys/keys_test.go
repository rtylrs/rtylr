package keys

import (
	"path"
	"testing"
)

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

func TestSubscriptionPatternsDelimitOwnerID(t *testing.T) {
	patterns := SubscriptionPatterns(4)
	want := []string{"cache:subscription:4", "cache:subscription:4:*"}
	if len(patterns) != len(want) {
		t.Fatalf("SubscriptionPatterns(4) = %v, want %v", patterns, want)
	}
	for i := range want {
		if patterns[i] != want[i] {
			t.Fatalf("SubscriptionPatterns(4) = %v, want %v", patterns, want)
		}
	}

	// The wildcard covers the owner's derived keys but must not bleed into
	// neighbouring owner ids (4 vs 42) the way "cache:subscription:4*" would.
	matched, err := path.Match(SubscriptionPrefix(4), SubscriptionAccess(4))
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("prefix %q does not match %q", SubscriptionPrefix(4), SubscriptionAccess(4))
	}
	matched, err = path.Match(SubscriptionPrefix(4), Subscription(42))
	if err != nil {
		t.Fatal(err)
	}
	if matched {
		t.Fatalf("prefix %q leaked into owner 42's key", SubscriptionPrefix(4))
	}
}
