package manager_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/modules/pos/manager"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
	"github.com/voxire/onion/internal/testutil"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

func TestManagerPINStepUpUsesOrganizationVerifier(t *testing.T) {
	db := testutil.OpenDB(t)
	const businessID, ownerID, staffID = int64(9968401), int64(9968402), int64(9968403)
	testutil.GrantBusinessAccess(t, db, businessID, ownerID)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = ?", ownerID) })
	hash, err := passwordutils.HashPassword("Synthetic-Password-97")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hash, ownerID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM pos_manager_pins WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM pos_manager_pin_attempts WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM activity_log WHERE business_id = ?", businessID)
	})
	mux := http.NewServeMux()
	manager.NewModule(context.Background(), db).WithPasswordVerifier(scopes.NewReader(db).VerifyAccountPassword).RegisterAPI(mux)
	user := &auth.AuthUser{UserID: ownerID, BusinessID: businessID, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true, Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppPOS, Licensed: true, HasAccess: true, Permissions: []string{"pos:orders:view"}}}}}
	request := func(route string, body *posmanagerv1.SetPINRequest, want int) *httptest.ResponseRecorder {
		t.Helper()
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, route, bytes.NewReader(encoded))
		req = req.WithContext(auth.WithAuthUser(req.Context(), user))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		if response.Code != want {
			t.Fatalf("%s status=%d want=%d: %s", route, response.Code, want, response.Body.String())
		}
		return response
	}
	request("/pos/manager/v1/set-pin", &posmanagerv1.SetPINRequest{StaffId: staffID, Pin: "582749", CurrentPassword: "Synthetic-Password-97"}, http.StatusOK)
	var stored string
	if err := db.QueryRow("SELECT pin_hash FROM pos_manager_pins WHERE business_id = ? AND staff_id = ?", businessID, staffID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == "582749" || !passwordutils.CheckPINHash("582749", stored) {
		t.Fatal("manager PIN was not stored as a usable hash")
	}
	request("/pos/manager/v1/set-pin", &posmanagerv1.SetPINRequest{StaffId: staffID, Pin: "963827", CurrentPassword: "wrong-password"}, http.StatusUnauthorized)
	var after string
	if err := db.QueryRow("SELECT pin_hash FROM pos_manager_pins WHERE business_id = ? AND staff_id = ?", businessID, staffID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != stored {
		t.Fatal("wrong account password changed the manager credential")
	}
	var failures int
	if err := db.QueryRow("SELECT failures FROM pos_manager_pin_attempts WHERE business_id = ?", businessID).Scan(&failures); err != nil {
		t.Fatal(err)
	}
	if failures != 1 {
		t.Fatalf("failed step-up attempts=%d, want 1", failures)
	}
}
