package featuresupport

import "testing"

func TestAnonymousRoutesMatchGeneratedFeatureRoute(t *testing.T) {
	routes := (&Module{}).GetAnonymousRoutes()
	if len(routes) != 1 || routes[0] != "/hub/feature/v1/features" {
		t.Fatalf("unexpected anonymous feature routes: %v", routes)
	}
}
