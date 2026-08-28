package business

import "testing"

func TestPublicBusinessRoutesAreAnonymous(t *testing.T) {
	want := map[string]struct{}{
		"/hub/business/v1/businesses/by-slug":     {},
		"/hub/business/v1/businesses/public":      {},
		"/hub/business/v1/businesses/name-exists": {},
	}
	got := (&Module{}).GetAnonymousRoutes()
	if len(got) != len(want) {
		t.Fatalf("anonymous business route count = %d, want %d", len(got), len(want))
	}
	for _, route := range got {
		if _, ok := want[route]; !ok {
			t.Errorf("unexpected anonymous business route %q", route)
		}
	}
}
