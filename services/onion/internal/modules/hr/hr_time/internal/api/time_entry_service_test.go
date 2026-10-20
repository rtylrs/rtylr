package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

// Finance's invoice screens read billable hours through the hr.time_entry.v1
// contract and stamp them once invoiced. Nothing served that contract, so the
// exact request below was a 404.
func TestTimeEntryContract_ServesFinanceBillableTimeOverHTTP(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9130001
	const userID int64 = 9131
	if _, err := db.ExecContext(ctx, "INSERT INTO businesses (id, name, owner_id) VALUES (?, 'HR time contract', ?)", businessID, userID); err != nil {
		t.Fatalf("seed business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
	})
	repo := infrastructure.NewRepository(db)
	if _, err := repo.CreateTimeEntry(ctx, businessID, domain.TimeEntryRecord{
		EmployeeID: 9132, Date: "2026-10-01", DurationMinutes: 90, IsBillable: true,
		Status: domain.EntryStatusApproved, Description: "Client workshop",
	}); err != nil {
		t.Fatalf("seed billable entry: %v", err)
	}
	if _, err := repo.CreateTimeEntry(ctx, businessID, domain.TimeEntryRecord{
		EmployeeID: 9132, Date: "2026-10-02", DurationMinutes: 30, IsBillable: false,
		Status: domain.EntryStatusApproved, Description: "Internal",
	}); err != nil {
		t.Fatalf("seed internal entry: %v", err)
	}

	service := application.NewService(ctx, repo, nil, w2APIPermissionProvider{employeeID: 9132, allowed: true})
	mux := http.NewServeMux()
	if err := hrtimeentryv1.RegisterTimeEntryServiceServer(NewTimeEntryAPI(service), hrtimeentryv1.WithMux(mux)); err != nil {
		t.Fatalf("register: %v", err)
	}
	authed := w2AuthenticatedContext(userID, businessID, "")
	send := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(authed)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	list := send(http.MethodGet, "/hr/time-entry/v1/time-entries?from_date=2026-07-22&to_date=2026-10-20&is_billable=true&status=approved&limit=200&offset=0", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list = %d %s", list.Code, list.Body.String())
	}
	var listed struct {
		Entries []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list %s: %v", list.Body.String(), err)
	}
	if len(listed.Entries) != 1 || listed.Entries[0].Description != "Client workshop" {
		t.Fatalf("listed %+v, want only the billable approved entry", listed.Entries)
	}

	marked := send(http.MethodPatch, "/hr/time-entry/v1/time-entries/mark-invoiced",
		`{"entry_ids":["`+listed.Entries[0].ID+`"],"invoice_id":"77"}`)
	if marked.Code == http.StatusNotFound && strings.Contains(marked.Body.String(), "page not found") {
		t.Fatalf("mark-invoiced is still unrouted: %s", marked.Body.String())
	}
}
