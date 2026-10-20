package main

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/server/http"
)

type parityListRequest struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
}
type parityListResponse struct {
	Search string   `json:"search"`
	Limit  int      `json:"limit"`
	Items  []string `json:"items"`
	Total  int      `json:"total"`
}
type parityList struct{}

func (parityList) Serve(_ context.Context, req *parityListRequest) (parityListResponse, error) {
	return parityListResponse{Search: req.Search, Limit: req.Limit}, nil
}

type parityUpdateRequest struct {
	Name string `json:"name"`
}
type parityUpdateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type parityUpdate struct{}

func (parityUpdate) Validate(_ context.Context, req *parityUpdateRequest) error {
	if req.Name == "" {
		return onekiterrors.New(stdhttp.StatusBadRequest, "name is required")
	}
	return nil
}

func (parityUpdate) Serve(ctx context.Context, req *parityUpdateRequest) (parityUpdateResponse, error) {
	if req.Name == "boom" {
		return parityUpdateResponse{}, onekiterrors.New(stdhttp.StatusConflict, "already exists")
	}
	return parityUpdateResponse{ID: http.PathParam(ctx, "id"), Name: req.Name}, nil
}

type parityDeleteRequest struct {
	ID int64 `json:"id"`
}
type parityDeleteResponse struct {
	Deleted int64 `json:"deleted"`
}
type parityDelete struct{}

func (parityDelete) Serve(_ context.Context, req *parityDeleteRequest) (parityDeleteResponse, error) {
	return parityDeleteResponse{Deleted: req.ID}, nil
}

// A route served through its compat contract must answer exactly as it did
// when mounted straight from the legacy table: same status, same body.
func TestCompatContractsAnswerExactlyLikeTheLegacyRoutes(t *testing.T) {
	ctx := context.Background()
	routes := http.NewLegacyRoutes()
	http.RegisterGET[parityListRequest, parityListResponse](ctx, routes, []string{"/crm/customers"}, parityList{})
	http.RegisterPUTWithValidation[parityUpdateRequest, parityUpdateResponse](ctx, routes, []string{"/crm/automations/{id}"}, parityUpdate{})
	http.RegisterDELETE[parityDeleteRequest, parityDeleteResponse](ctx, routes, []string{"/crm/customers"}, parityDelete{})

	legacy := stdhttp.NewServeMux()
	routes.Mount(legacy, nil)
	contract := stdhttp.NewServeMux()
	registerCompatContracts(contract, routes)
	routes.Mount(contract, compatRoutes().Contains)

	cases := []struct {
		method, target, body string
		// statusOnly marks the documented differences: the contract's own
		// 400 for a malformed typed query parameter words its message
		// differently, and it reads only the first of two concatenated JSON
		// bodies where the legacy helper rejected the request.
		statusOnly, lenient bool
	}{
		{stdhttp.MethodGet, "/crm/customers?search=acme&limit=5", "", false, false},
		{stdhttp.MethodGet, "/crm/customers?limit=notanumber", "", true, false},
		{stdhttp.MethodPut, "/crm/automations/42", `{"name":"Rana"}`, false, false},
		{stdhttp.MethodPut, "/crm/automations/42", `{"name":""}`, false, false},
		{stdhttp.MethodPut, "/crm/automations/42", `{"name":"boom"}`, false, false},
		{stdhttp.MethodPut, "/crm/automations/42", `{"name":"a"}{"name":"b"}`, false, true},
		{stdhttp.MethodDelete, "/crm/customers", `{"id":7}`, false, false},
		{stdhttp.MethodDelete, "/crm/customers?id=9", "", false, false},
	}
	for _, tc := range cases {
		do := func(mux *stdhttp.ServeMux) *httptest.ResponseRecorder {
			var body *strings.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tc.method, tc.target, body)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			return rec
		}
		want, got := do(legacy), do(contract)
		if tc.lenient {
			if got.Code != stdhttp.StatusOK {
				t.Errorf("%s %s: contract should accept the first JSON value, got %d", tc.method, tc.target, got.Code)
			}
			continue
		}
		if tc.statusOnly {
			if got.Code != want.Code {
				t.Errorf("%s %s: status legacy %d, contract %d", tc.method, tc.target, want.Code, got.Code)
			}
			continue
		}
		if got.Code != want.Code || strings.TrimSpace(got.Body.String()) != strings.TrimSpace(want.Body.String()) {
			t.Errorf("%s %s %s:\n legacy   %d %s\n contract %d %s", tc.method, tc.target, tc.body,
				want.Code, strings.TrimSpace(want.Body.String()), got.Code, strings.TrimSpace(got.Body.String()))
		}
	}
}

// Routes behind a guard or a rate limit keep it when a contract serves them:
// @http methods run the original handler with its route middlewares.
func TestCompatContractsKeepRouteMiddlewares(t *testing.T) {
	ctx := context.Background()
	routes := http.NewLegacyRoutes()
	guard := func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			if r.Header.Get("Authorization") != "Bearer admin" {
				w.WriteHeader(stdhttp.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	admin := routes.With(guard)
	admin.Get("/voxhire/admin/me", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte(`{"admin":true}`))
	})
	http.RegisterPOSTWithValidation[parityUpdateRequest, parityUpdateResponse](ctx, admin, []string{"/voxhire/admin/careers"}, parityUpdate{})

	mux := stdhttp.NewServeMux()
	registerCompatContracts(mux, routes)

	for _, tc := range []struct {
		method, target, body, auth string
		want                       int
	}{
		{stdhttp.MethodGet, "/voxhire/admin/me", "", "", stdhttp.StatusUnauthorized},
		{stdhttp.MethodGet, "/voxhire/admin/me", "", "Bearer admin", stdhttp.StatusOK},
		{stdhttp.MethodPost, "/voxhire/admin/careers", `{"name":"Engineer"}`, "", stdhttp.StatusUnauthorized},
		{stdhttp.MethodPost, "/voxhire/admin/careers", `{"name":"Engineer"}`, "Bearer admin", stdhttp.StatusOK},
	} {
		req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
		if tc.auth != "" {
			req.Header.Set("Authorization", tc.auth)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s %s auth=%q: status %d, want %d (%s)", tc.method, tc.target, tc.auth, rec.Code, tc.want, rec.Body)
		}
	}
}
