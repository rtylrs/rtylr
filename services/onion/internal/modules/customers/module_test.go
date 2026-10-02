package customers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authpkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	customerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
)

func TestGeneratedCustomerRoutesRequireAuthorization(t *testing.T) {
	mux := http.NewServeMux()
	(&Module{}).RegisterAPI(mux)
	for _, tc := range []struct{ method, path, body string }{
		{"GET", "/crm/customer/v1/customers", ""},
		{"POST", "/crm/customer/v1/customers", `{"first_name":"Test"}`},
		{"GET", "/crm/customer/v1/customers/1", ""},
		{"PUT", "/crm/customer/v1/customers/1", `{"first_name":"Test"}`},
		{"DELETE", "/crm/customer/v1/customers/1", ""},
		{"GET", "/crm/dashboard/v1/dashboard", ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			mux.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestGeneratedCustomerCRUD(t *testing.T) {
	db := testutil.OpenDB(t)
	const businessID int64 = 741201
	const otherBusinessID int64 = 741202
	testutil.GrantBusinessAccess(t, db, businessID, businessID)
	testutil.GrantBusinessAccess(t, db, otherBusinessID, otherBusinessID)
	mux := http.NewServeMux()
	NewModule(context.Background(), db).RegisterAPI(mux)
	user := func(business int64, permissions ...string) *authpkg.AuthUser {
		return &authpkg.AuthUser{BusinessID: business, UserID: business, EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppCRM, Licensed: true, HasAccess: true, Permissions: permissions}}}}
	}
	owner := user(businessID, "crm:customers:view", "crm:customers:create", "crm:customers:edit", "crm:customers:delete")
	call := func(as *authpkg.AuthUser, method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request = request.WithContext(authpkg.WithAuthUser(request.Context(), as))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != status {
			t.Fatalf("%s %s: status = %d, want %d: %s", method, path, response.Code, status, response.Body.String())
		}
		return response
	}
	created := call(owner, "POST", "/crm/customer/v1/customers", `{"first_name":"Portal","last_name":"Fixture","email":"portal-crm@example.test","phone":"+15555550117","consent":{"email":true}}`, 200)
	var result customerv1.CreateCustomerResponse
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.CustomerId <= 0 {
		t.Fatalf("invalid customer id: %d", result.CustomerId)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM crm_customers WHERE id = ?", result.CustomerId) })
	path := fmt.Sprintf("/crm/customer/v1/customers/%d", result.CustomerId)
	list := call(owner, "GET", "/crm/customer/v1/customers?search=Portal&page=1&page_size=20", "", 200)
	if !strings.Contains(list.Body.String(), "portal-crm@example.test") {
		t.Fatalf("created customer missing from list: %s", list.Body.String())
	}
	call(user(businessID, "crm:customers:view"), "PUT", path, `{"first_name":"Forbidden"}`, 403)
	call(user(otherBusinessID, "crm:customers:view"), "GET", path, "", 404)
	call(owner, "PUT", path, `{"first_name":"Updated","last_name":"Fixture","email":"portal-crm@example.test","phone":"+15555550117","consent":{"email":true}}`, 200)
	detail := call(owner, "GET", path, "", 200)
	if !strings.Contains(detail.Body.String(), `"first_name":"Updated"`) {
		t.Fatalf("update not persisted: %s", detail.Body.String())
	}
	call(owner, "GET", "/crm/dashboard/v1/dashboard", "", 200)
	call(owner, "DELETE", path, "", 200)
	call(owner, "GET", path, "", 404)
}
