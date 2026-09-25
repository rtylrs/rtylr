package tenant

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestResolveBusinessIDFailsClosedWithoutDatabase(t *testing.T) {
	resolved, err := resolveBusinessID(context.Background(), nil, 42, 7)

	if resolved != 0 {
		t.Fatalf("resolved business id = %d, want 0", resolved)
	}
	if !errors.Is(err, ErrNoBusinessScope) {
		t.Fatalf("error = %v, want ErrNoBusinessScope", err)
	}
}

func TestResolveBusinessIDRequiresAuthenticatedUser(t *testing.T) {
	resolved, err := resolveBusinessID(context.Background(), nil, 42, 0)

	if resolved != 0 {
		t.Fatalf("resolved business id = %d, want 0", resolved)
	}
	if !errors.Is(err, ErrNoBusinessScope) {
		t.Fatalf("error = %v, want ErrNoBusinessScope", err)
	}
}

func TestBusinessAccessQueryChecksActiveBusinessMembership(t *testing.T) {
	query := strings.ToLower(businessAccessQuery)

	for _, expected := range []string{
		"coalesce(b.is_active, true) = true",
		"b.owner_id = ?",
		"from user_business_access uba",
		"uba.user_id = ?",
	} {
		if !strings.Contains(query, expected) {
			t.Errorf("business access query missing %q", expected)
		}
	}
}

func TestBusinessAccessQueryRecognizesNormalizedAndLegacyMembership(t *testing.T) {
	query := strings.ToLower(businessAccessQuery)

	for _, expected := range []string{
		"from authorization_business_members abm",
		"abm.user_id = ?",
		"abm.status = 'active'",
		"from user_app_permissions uap",
		"uap.user_id = ?",
		"uap.is_active = true",
	} {
		if !strings.Contains(query, expected) {
			t.Errorf("business access query missing %q", expected)
		}
	}
}
