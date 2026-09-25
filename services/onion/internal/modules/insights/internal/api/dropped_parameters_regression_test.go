package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	insightssettingv1 "github.com/voxire/voxie/gen/go/insights/setting/v1"
)

func TestScheduledReportPresenceMiddlewareDistinguishesOmittedAndFalse(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantNil    bool
		wantActive bool
	}{
		{name: "omitted", body: `{}`, wantNil: true},
		{name: "explicit false", body: `{"is_active":false}`, wantActive: false},
		{name: "explicit true", body: `{"is_active":true}`, wantActive: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got *bool
			handler := ScheduledReportPresenceMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				var fields struct {
					IsActive bool `json:"is_active"`
				}
				if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				got = scheduledReportIsActive(r.Context(), fields.IsActive)
			}))
			req := httptest.NewRequest(http.MethodPost, "/insights/report/v1/scheduled", bytes.NewBufferString(tc.body))
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if tc.wantNil {
				if got != nil {
					t.Fatalf("is_active presence = %v, want nil", *got)
				}
				return
			}
			if got == nil || *got != tc.wantActive {
				t.Fatalf("is_active = %v, want %v", got, tc.wantActive)
			}
		})
	}
}

func TestScheduledReportPresenceMiddlewareRestoresBody(t *testing.T) {
	const body = `{"template_slug":"revenue-monthly","frequency":"monthly","channel":"email","recipient_email":"a@example.com"}`
	var decoded map[string]any
	handler := ScheduledReportPresenceMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read restored body: %v", err)
		}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("decode restored body: %v", err)
		}
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/insights/report/v1/scheduled", bytes.NewBufferString(body)))
	if decoded["template_slug"] != "revenue-monthly" {
		t.Fatalf("restored body = %#v", decoded)
	}
}

func TestSettingsToProtoPersistsAllContractFields(t *testing.T) {
	got := settingsToProto(application.SettingsPayload{
		DefaultPeriod: "this_quarter", DefaultCurrency: "EUR", ComparePrevious: true,
		DashboardWidgets: []string{"revenue"}, FavoriteReports: []string{"weekly-revenue"},
	})
	want := &insightssettingv1.InsightsSettings{
		DefaultPeriod: "this_quarter", Currency: "EUR", ComparePrevious: true,
		DashboardWidgets: []string{"revenue"}, FavoriteReports: []string{"weekly-revenue"},
	}
	if got.DefaultPeriod != want.DefaultPeriod || got.Currency != want.Currency || got.ComparePrevious != want.ComparePrevious {
		t.Fatalf("scalar settings = %+v, want %+v", got, want)
	}
	if len(got.DashboardWidgets) != 1 || got.DashboardWidgets[0] != "revenue" || len(got.FavoriteReports) != 1 || got.FavoriteReports[0] != "weekly-revenue" {
		t.Fatalf("list settings = %+v, want %+v", got, want)
	}
}

func TestPaginationResponseIncludesPageMetadata(t *testing.T) {
	got := paginationResponse(2, 10, 23)
	if got.Page != 2 || got.Limit != 10 || got.Total != 23 || got.TotalPages != 3 || !got.HasNext || !got.HasPrev {
		t.Fatalf("pagination = %+v", got)
	}
}
