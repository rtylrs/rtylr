package websitelead

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
)

type fakeRepo struct {
	leads       []domain.LeadInput
	subscribers []domain.SubscriberInput
}

func (f *fakeRepo) Create(_ context.Context, in domain.LeadInput) (int64, error) {
	f.leads = append(f.leads, in)
	return int64(len(f.leads)), nil
}

func (f *fakeRepo) CreateSubscriber(_ context.Context, in domain.SubscriberInput) (int64, error) {
	f.subscribers = append(f.subscribers, in)
	return int64(len(f.subscribers)), nil
}

func newServer(t *testing.T) (*http.ServeMux, *fakeRepo) {
	t.Helper()
	repo := &fakeRepo{}
	m := &Module{service: application.NewService(context.Background(), repo, Notify{})}
	mux := http.NewServeMux()
	m.RegisterAPI(mux)
	return mux, repo
}

func post(mux *http.ServeMux, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.7:1234"
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// The website posts to these exact paths, so the contract must keep them.
func TestRoutesKeepTheWebsitePaths(t *testing.T) {
	mux, repo := newServer(t)
	for _, tc := range []struct{ path, body string }{
		{"/quote", `{"name":"Ada","email":"ada@example.com","service":"pos","timeline":"Q1","project":"New store"}`},
		{"/contact", `{"name":"Ada","email":"ada@example.com"}`},
		{"/newsletter", `{"email":"  Ada@Example.com ","source":"intel_drop"}`},
		{"/analytics/event", `{"event":"view","job_id":42}`},
	} {
		if rec := post(mux, tc.path, tc.body); rec.Code != http.StatusOK {
			t.Fatalf("POST %s = %d %s, want 200", tc.path, rec.Code, rec.Body.String())
		}
	}
	if len(repo.leads) != 2 || repo.leads[0].Project != "Timeline: Q1\n\nNew store" || repo.leads[1].Service != "contact" {
		t.Fatalf("leads = %+v", repo.leads)
	}
	if len(repo.subscribers) != 1 || repo.subscribers[0].Email != "ada@example.com" || repo.subscribers[0].Language != "en" {
		t.Fatalf("subscriber not normalised: %+v", repo.subscribers)
	}
}

func TestValidationRejectsBadInput(t *testing.T) {
	mux, repo := newServer(t)
	for _, tc := range []struct{ name, path, body, want string }{
		{"quote needs a service", "/quote", `{"name":"Ada","email":"ada@example.com"}`, ""},
		{"contact needs a name", "/contact", `{"email":"ada@example.com"}`, ""},
		{"newsletter email", "/newsletter", `{"email":"not-an-email","source":"intel_drop"}`, "email is invalid"},
		{"newsletter source", "/newsletter", `{"email":"ada@example.com","source":"web3forms"}`, "source is invalid"},
		{"newsletter language", "/newsletter", `{"email":"ada@example.com","source":"intel_drop","language":"fr"}`, "language is invalid"},
		{"newsletter title length", "/newsletter", `{"email":"ada@example.com","source":"intel_drop","post_title":"` + strings.Repeat("t", 256) + `"}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(mux, tc.path, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body.String())
			}
			if tc.want != "" && !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body %q does not mention %q", rec.Body.String(), tc.want)
			}
		})
	}
	if len(repo.leads)+len(repo.subscribers) != 0 {
		t.Fatal("an invalid request was stored")
	}
}

func TestRateLimitComesFromTheContract(t *testing.T) {
	mux, _ := newServer(t)
	body := `{"name":"Ada","email":"ada@example.com"}`
	for i := 0; i < 10; i++ {
		if rec := post(mux, "/contact", body); rec.Code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200", i+1, rec.Code)
		}
	}
	if rec := post(mux, "/contact", body); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("11th request = %d, want 429 from @meta(rate_limit, 10/m)", rec.Code)
	}
	// Budgets are per route: another route still answers.
	if rec := post(mux, "/analytics/event", `{"event":"view"}`); rec.Code != http.StatusOK {
		t.Fatalf("analytics after contact limit = %d, want 200", rec.Code)
	}
}
