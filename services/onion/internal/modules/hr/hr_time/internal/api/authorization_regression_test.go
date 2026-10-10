package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

type w2APIPermissionProvider struct {
	employeeID int64
	allowed    bool
}

func (p w2APIPermissionProvider) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return p.employeeID, nil
}

func (p w2APIPermissionProvider) GetEmployeeBusinessID(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

func (p w2APIPermissionProvider) HasPermission(context.Context, int64, int64, string) (bool, error) {
	return p.allowed, nil
}

func (p w2APIPermissionProvider) HasRole(context.Context, int64, int64, int64) (bool, error) {
	return false, nil
}

func w2AuthenticatedContext(userID, businessID int64, urlID string) context.Context {
	device := &authPkg.Device{User: &authPkg.AuthUser{UserID: userID, BusinessID: businessID}}
	ctx := device.Ctx(context.Background())
	if urlID == "" {
		return ctx
	}
	routeCtx := map[string]string{}
	routeCtx["id"] = (urlID)
	return serverHttp.WithPathParams(ctx, routeCtx)
}

func w2APIStatus(err error) int {
	var httpErr *onekithttp.Error
	if errors.As(err, &httpErr) {
		return int(httpErr.GetStatusCode())
	}
	return 0
}

func TestDeleteTimeEntryHandler_CrossEmployeeRefusalIsForbidden(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9120001
	const userID int64 = 9121
	const callerEmployeeID int64 = 9122
	const ownerEmployeeID int64 = 9123

	if _, err := db.ExecContext(ctx, "INSERT INTO businesses (id, name, owner_id) VALUES (?, 'W2 HR Time Auth', ?)", businessID, userID); err != nil {
		t.Fatalf("seed business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
	})
	repo := infrastructure.NewRepository(db)
	entry, err := repo.CreateTimeEntry(ctx, businessID, domain.TimeEntryRecord{
		EmployeeID:      ownerEmployeeID,
		Date:            "2026-11-06",
		DurationMinutes: 30,
	})
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	service := application.NewService(ctx, infrastructure.NewRepository(db), nil, w2APIPermissionProvider{employeeID: callerEmployeeID})
	handler := &DeleteTimeEntryHandler{Service: service}
	_, err = handler.Serve(w2AuthenticatedContext(userID, businessID, strconv.FormatInt(entry.ID, 10)), &DeleteTimeEntryRequest{})
	if err == nil {
		t.Fatal("cross-employee delete unexpectedly succeeded")
	}
	if got := w2APIStatus(err); got != http.StatusForbidden {
		t.Fatalf("cross-employee delete status = %d, want %d (err: %v)", got, http.StatusForbidden, err)
	}
}
