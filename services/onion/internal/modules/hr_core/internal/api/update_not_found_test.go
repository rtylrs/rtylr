package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	onekithttp "github.com/corezio/onekit/http"
	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	workschedulev1 "github.com/voxire/voxie/gen/go/hr/work_schedule/v1"
)

// hrUpdateContext builds the request context an HR administration write runs
// under: an authenticated owner holding hr:roles:manage with a resolved HR app
// scope, plus the chi route param the handler reads.
func hrUpdateContext(businessID, userID int64, param string, id int64) context.Context {
	return hrRouteContext(businessID, userID, param, strconv.FormatInt(id, 10))
}

func hrRouteContext(businessID, userID int64, param, value string) context.Context {
	user := &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		Role:       "owner",
		ProvisioningStatus: &authorization.ProvisioningStatus{
			State:         authorization.ProvisioningActive,
			EmailVerified: true,
		},
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           businessID,
			DashRole:             "owner",
			IsOwner:              true,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:roles:manage"},
				Scopes:      []authorization.Scope{{Type: "location", ID: "1"}},
			}},
		},
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(param, value)
	return authPkg.WithAuthUser(context.WithValue(context.Background(), chi.RouteCtxKey, rctx), user)
}

func requireNotFound(t *testing.T, resp any, err error) {
	t.Helper()

	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusNotFound {
		t.Fatalf("expected 404, got response=%+v err=%v", resp, err)
	}
}

// Regression: the update handlers echoed the request back with 200 for an id
// that matched no row (nonexistent, stale, or another tenant's). The write must
// report 404 instead of pretending to have stored the record.
func TestUpdateLocationHandler_MissingID_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(982001)
		businessID = int64(982002)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &UpdateLocationHandler{Service: application.NewService(ctx, db)}
	resp, err := handler.Serve(
		hrUpdateContext(businessID, userID, "id", 999999982),
		&UpdateLocationRequest{Name: "Ghost Location"},
	)
	requireNotFound(t, resp, err)
}

func TestUpdateHRCoreDepartmentHandler_MissingID_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(982011)
		businessID = int64(982012)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &UpdateHRCoreDepartmentHandler{Service: application.NewService(ctx, db)}
	resp, err := handler.Serve(
		hrUpdateContext(businessID, userID, "id", 999999982),
		&UpdateHRCoreDepartmentRequest{Name: "Ghost Department"},
	)
	requireNotFound(t, resp, err)
}

func TestUpdateRoleHandler_MissingID_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(982021)
		businessID = int64(982022)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &UpdateRoleHandler{Service: application.NewService(ctx, db)}
	resp, err := handler.Serve(
		hrUpdateContext(businessID, userID, "id", 999999982),
		&UpdateRoleRequest{Name: "Ghost Role", Permissions: `{}`},
	)
	requireNotFound(t, resp, err)
}

func TestUpdateWorkSchedule_MissingID_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(982031)
		businessID = int64(982032)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	api := NewWorkScheduleAPI(application.NewService(ctx, db))
	name := "Ghost Schedule"
	resp, err := api.UpdateWorkSchedule(
		hrUpdateContext(businessID, userID, "id", 999999982),
		&workschedulev1.UpdateWorkScheduleRequest{
			Id:   999999982,
			Body: &workschedulev1.UpdateWorkScheduleBody{Name: &name},
		},
	)
	requireNotFound(t, resp, err)
}

// Regression: an unknown step_key surfaced sql.ErrNoRows as a 500 instead of a
// 404.
func TestCompleteSetupStepHandler_UnknownKey_Returns404(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		userID     = int64(982041)
		businessID = int64(982042)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	handler := &CompleteSetupStepHandler{Service: application.NewService(ctx, db)}
	resp, err := handler.Serve(
		hrRouteContext(businessID, userID, "step_key", "no-such-step-key"),
		&CompleteSetupStepRequest{},
	)
	requireNotFound(t, resp, err)
}
