package websitelead

import (
	"sort"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/api"
)

// TestAnonymousRoutesCoverPublicHandlers asserts that every route this
// module actually registers as public (SubmitRoutes, ContactRoutes,
// NewsletterRoutes, AnalyticsEventRoutes) appears in GetAnonymousRoutes, and
// that GetAnonymousRoutes lists nothing extra. A zero-value Module is safe
// here because GetAnonymousRoutes does not touch the unexported service
// field.
func TestAnonymousRoutesCoverPublicHandlers(t *testing.T) {
	var want []string
	want = append(want, api.SubmitRoutes()...)
	want = append(want, api.ContactRoutes()...)
	want = append(want, api.NewsletterRoutes()...)
	want = append(want, api.AnalyticsEventRoutes()...)

	got := (&Module{}).GetAnonymousRoutes()

	if !sameSet(want, got) {
		sortedWant := append([]string{}, want...)
		sortedGot := append([]string{}, got...)
		sort.Strings(sortedWant)
		sort.Strings(sortedGot)
		t.Fatalf("anonymous routes mismatch\nwant: %v\ngot:  %v", sortedWant, sortedGot)
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := append([]string{}, a...)
	sortedB := append([]string{}, b...)
	sort.Strings(sortedA)
	sort.Strings(sortedB)
	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}
	return true
}
