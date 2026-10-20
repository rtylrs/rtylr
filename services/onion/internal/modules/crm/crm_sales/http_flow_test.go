package crmsales_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	crmsales "github.com/voxire/onion/internal/modules/crm/crm_sales"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	authpkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

// salesServer mounts the module's generated routes behind a stand-in for the
// auth middleware: every request runs as the fixture business's owner.
func salesServer(t *testing.T) *httptest.Server {
	t.Helper()
	db := testutil.OpenDB(t)
	module := crmsales.NewModule(context.Background(), db).WithMembershipReader(scopes.NewReader(db).IsBusinessUser)
	mux := http.NewServeMux()
	module.RegisterAPI(mux)
	owner := &authpkg.AuthUser{
		UserID:     testutil.OwnerID,
		BusinessID: testutil.OwnerID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: testutil.OwnerID,
			IsOwner:    true,
			Apps:       []authorization.EffectiveAppAccess{{AppID: authorization.AppCRM, Licensed: true, HasAccess: true}},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r.WithContext(authpkg.WithAuthUser(r.Context(), owner)))
	}))
	t.Cleanup(server.Close)
	return server
}

func call(t *testing.T, server *httptest.Server, method, path string, body any, want int) map[string]any {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatalf("encode %s %s: %v", method, path, err)
		}
	}
	req, err := http.NewRequest(method, server.URL+path, &payload)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(res.Body).Decode(&out)
	if res.StatusCode != want {
		t.Fatalf("%s %s: status %d, want %d: %v", method, path, res.StatusCode, want, out)
	}
	return out
}

func object(t *testing.T, value any) map[string]any {
	t.Helper()
	node, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%v is not an object", value)
	}
	return node
}

func array(t *testing.T, value any) []any {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("%v is not an array", value)
	}
	return items
}

func field(t *testing.T, node map[string]any, path ...string) any {
	t.Helper()
	var current any = node
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("%v: %q is not inside an object in %v", path, key, node)
		}
		current = object[key]
	}
	return current
}

// The whole prospecting flow over HTTP, wire format included: create a lead,
// call it, schedule a follow-up, convert it, win the deal with a contract,
// and import a list.
func TestSalesFlowOverHTTP(t *testing.T) {
	server := salesServer(t)
	company := fmt.Sprintf("Http Co %d", time.Now().UnixNano())

	created := call(t, server, http.MethodPost, "/crm/sales/v1/leads", map[string]any{
		"lead": map[string]any{
			"first_name": "Rana", "last_name": "Haddad", "phone": "+961 70 123 456",
			"email": "Rana@Example.com", "company_name": company, "value": "250.00", "billing": "monthly",
		},
	}, http.StatusOK)
	leadID := fmt.Sprint(field(t, created, "lead", "id"))
	if field(t, created, "lead", "email") != "rana@example.com" || field(t, created, "lead", "status") != "new" {
		t.Fatalf("created lead = %v", created)
	}

	call(t, server, http.MethodPost, "/crm/sales/v1/activities", map[string]any{
		"lead_id": leadID, "type": "call", "outcome": "callback", "notes": "Call Thursday",
		"follow_up": map[string]any{"type": "call", "title": "Call back", "due_date": "2026-10-22"},
	}, http.StatusOK)

	list := call(t, server, http.MethodGet, "/crm/sales/v1/leads?statuses=connected&search="+"70123456"+"&today=2026-10-20", nil, http.StatusOK)
	leads := array(t, field(t, list, "leads"))
	if len(leads) != 1 || field(t, list, "total") != float64(1) {
		t.Fatalf("searching the called lead by phone digits: %v", list)
	}
	if follow := field(t, object(t, leads[0]), "next_follow_up"); follow != "2026-10-22" {
		t.Fatalf("next_follow_up = %v", follow)
	}

	tasks := call(t, server, http.MethodGet, "/crm/sales/v1/tasks?lead_id="+leadID+"&today=2026-10-20", nil, http.StatusOK)
	if items := array(t, field(t, tasks, "tasks")); len(items) != 1 || field(t, object(t, items[0]), "phone") != "+961 70 123 456" {
		t.Fatalf("the follow-up task should list the lead's phone: %v", tasks)
	}

	converted := call(t, server, http.MethodPut, "/crm/sales/v1/leads/"+leadID+"/conversion", map[string]any{"create_deal": true}, http.StatusOK)
	dealID := fmt.Sprint(field(t, converted, "deal", "id"))
	if field(t, converted, "lead", "status") != "qualified" || field(t, converted, "deal", "amount") != "3000.00" {
		t.Fatalf("converted = %v", converted)
	}
	call(t, server, http.MethodPut, "/crm/sales/v1/leads/"+leadID+"/conversion", map[string]any{"create_deal": true}, http.StatusConflict)

	board := call(t, server, http.MethodGet, "/crm/sales/v1/deals/board?search="+url.QueryEscape(company), nil, http.StatusOK)
	columns := array(t, field(t, board, "columns"))
	if len(columns) != 6 || field(t, object(t, columns[0]), "deals") != float64(1) {
		t.Fatalf("board = %v", board)
	}

	won := call(t, server, http.MethodPut, "/crm/sales/v1/deals/"+dealID+"/win", map[string]any{"contract": map[string]any{}}, http.StatusOK)
	if field(t, won, "deal", "status") != "won" || field(t, won, "contract", "deal_id") != dealID || field(t, won, "contract", "value") != "250.00" {
		t.Fatalf("won = %v", won)
	}
	deal := call(t, server, http.MethodGet, "/crm/sales/v1/deals/"+dealID, nil, http.StatusOK)
	if field(t, deal, "contract", "plan") != company {
		t.Fatalf("the deal should carry its contract: %v", deal)
	}
	call(t, server, http.MethodDelete, "/crm/sales/v1/deals/"+dealID, nil, http.StatusConflict)

	started := call(t, server, http.MethodPost, "/crm/sales/v1/lead-imports", map[string]any{"file_name": "list.csv", "total_rows": 2}, http.StatusOK)
	importID := fmt.Sprint(field(t, started, "import", "id"))
	rows := call(t, server, http.MethodPost, "/crm/sales/v1/lead-imports/"+importID+"/rows", map[string]any{
		"rows": []any{
			map[string]any{"row": 2, "lead": map[string]any{"company_name": company + " list", "phone": fmt.Sprintf("555 %07d", time.Now().UnixNano()%10000000)}},
			map[string]any{"row": 3, "lead": map[string]any{"email": "no-name@example.com"}},
		},
		"source": "Spring list",
	}, http.StatusOK)
	outcomes := map[string]string{}
	for _, result := range array(t, field(t, rows, "results")) {
		row := object(t, result)
		outcomes[fmt.Sprint(row["row"])] = fmt.Sprint(row["outcome"])
	}
	if outcomes["2"] != "created" || outcomes["3"] != "failed" {
		t.Fatalf("import results = %v", rows)
	}
	finished := call(t, server, http.MethodPut, "/crm/sales/v1/lead-imports/"+importID, nil, http.StatusOK)
	if field(t, finished, "import", "status") != "completed" || field(t, finished, "import", "failed") != float64(1) {
		t.Fatalf("finished = %v", finished)
	}

	call(t, server, http.MethodGet, "/crm/sales/v1/leads?sort=sideways", nil, http.StatusBadRequest)
	call(t, server, http.MethodGet, "/crm/sales/v1/leads/999999999", nil, http.StatusNotFound)
}
