package campaigns

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAPIRegistersGeneratedCampaignRoutes(t *testing.T) {
	module := &Module{}
	registrar, ok := any(module).(interface{ RegisterAPI(*http.ServeMux) })
	if !ok {
		t.Fatal("campaign module does not expose generated API registration")
	}
	mux := http.NewServeMux()
	registrar.RegisterAPI(mux)

	paths := []string{
		"/crm/campaign/v1/campaigns",
		"/crm/campaign/v1/campaigns/7",
		"/crm/campaign/v1/campaigns/stats?campaign_id=7",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			res := httptest.NewRecorder()
			mux.ServeHTTP(res, req)

			if res.Code == http.StatusNotFound {
				t.Fatalf("generated campaign route %q was not registered", path)
			}
			if res.Code != http.StatusForbidden {
				t.Fatalf("anonymous request to %q returned %d, want 403", path, res.Code)
			}
		})
	}
}
