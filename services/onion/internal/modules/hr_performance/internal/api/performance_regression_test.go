package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hrperformance "github.com/voxire/onion/internal/modules/hr_performance"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/api"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func performanceAPITestID(base int64) int64 {
	return base + time.Now().UnixNano()%100_000_000
}

func performanceAuthContext(userID, businessID int64, permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		Role:       "member",
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           businessID,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
				Scopes:      []authorization.Scope{{Type: "business", ID: fmt.Sprintf("%d", businessID)}},
			}},
		},
	})
}

func ensurePerformanceBusiness(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO businesses (id, name, owner_id, is_active)
		VALUES (?, ?, ?, TRUE)
		ON DUPLICATE KEY UPDATE is_active = TRUE`,
		businessID, fmt.Sprintf("performance-regression-%d", businessID), businessID); err != nil {
		t.Fatalf("create business: %v", err)
	}
}

func grantPerformanceBusiness(t *testing.T, db *sql.DB, userID, businessID int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO user_business_access (user_id, business_id, role)
		VALUES (?, ?, 'viewer')
		ON DUPLICATE KEY UPDATE role = 'viewer'`, userID, businessID); err != nil {
		t.Fatalf("grant business access: %v", err)
	}
}

func insertPerformanceAPIMployee(t *testing.T, db *sql.DB, businessID, employeeID int64, userID *int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO hr_employees
			(id, owner_id, business_id, user_id, first_name, last_name, email, status)
		VALUES (?, ?, ?, ?, 'Performance', 'API', ?, 'active')`,
		employeeID, businessID, businessID, userID,
		fmt.Sprintf("performance-api-%d@example.invalid", employeeID))
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}
}

func performanceHTTPStatus(err error) int {
	var statusErr interface{ HTTPStatusCode() int }
	if errors.As(err, &statusErr) {
		return statusErr.HTTPStatusCode()
	}
	return 0
}

func TestCreateCycleUsesResolvedBusinessAndRequiresManagePermission(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := performanceAPITestID(9_000_000_000)
	otherBusinessID := businessID + 1
	userID := otherBusinessID // deliberately differs from the selected business
	ensurePerformanceBusiness(t, db, businessID)
	ensurePerformanceBusiness(t, db, otherBusinessID)
	grantPerformanceBusiness(t, db, userID, businessID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?", userID, businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id IN (?, ?)", businessID, otherBusinessID)
	})

	service := application.NewService(ctx, db)
	handler := &api.CreateCycleHandler{Service: service}
	cycleRequest := &api.CreateCycleRequest{
		Name:        fmt.Sprintf("Business scoped cycle %d", businessID),
		CycleType:   domain.CycleTypeAdHoc,
		PeriodStart: "2031-01-01",
		PeriodEnd:   "2031-12-31",
	}
	created, err := handler.Serve(performanceAuthContext(userID, businessID, "hr:employees:manage"), cycleRequest)
	if err != nil {
		t.Fatalf("authorized cycle create: %v", err)
	}
	if created.Cycle == nil || created.Cycle.GetOwnerId() != businessID {
		t.Fatalf("created cycle owner = %+v, want business %d", created.Cycle, businessID)
	}
	var ownerID, storedBusinessID int64
	if err := db.QueryRowContext(ctx, `SELECT owner_id, business_id FROM hr_performance_review_cycles WHERE id = ?`, created.Cycle.GetId()).Scan(&ownerID, &storedBusinessID); err != nil {
		t.Fatalf("read created cycle: %v", err)
	}
	if ownerID != businessID || storedBusinessID != businessID {
		t.Fatalf("stored cycle scope = owner %d/business %d, want %d", ownerID, storedBusinessID, businessID)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_cycles WHERE id = ?", created.Cycle.GetId())
	})

	_, err = handler.Serve(performanceAuthContext(userID, businessID, "hr:employees:view"), &api.CreateCycleRequest{
		Name:        fmt.Sprintf("Viewer denied cycle %d", businessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2032-01-01",
		PeriodEnd:   "2032-12-31",
	})
	if status := performanceHTTPStatus(err); status != http.StatusForbidden {
		t.Fatalf("viewer cycle create status = %d, want 403 (err=%v)", status, err)
	}
}

func TestCreateGoalAuthorizesEmployeeSubject(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := performanceAPITestID(9_100_000_000)
	foreignBusinessID := businessID + 1
	userID := performanceAPITestID(9_200_000_000)
	employeeID := performanceAPITestID(9_300_000_000)
	peerEmployeeID := employeeID + 1
	foreignEmployeeID := employeeID + 2
	ensurePerformanceBusiness(t, db, businessID)
	ensurePerformanceBusiness(t, db, foreignBusinessID)
	grantPerformanceBusiness(t, db, userID, businessID)
	insertPerformanceAPIMployee(t, db, businessID, employeeID, &userID)
	insertPerformanceAPIMployee(t, db, businessID, peerEmployeeID, nil)
	insertPerformanceAPIMployee(t, db, foreignBusinessID, foreignEmployeeID, nil)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_goals WHERE business_id IN (?, ?)", businessID, foreignBusinessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_cycles WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?", userID, businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id IN (?, ?, ?)", employeeID, peerEmployeeID, foreignEmployeeID)
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id IN (?, ?)", businessID, foreignBusinessID)
	})

	repo := infrastructure.NewRepository(db)
	cycle, err := repo.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
		Name:        fmt.Sprintf("Goal subject cycle %d", businessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2033-01-01",
		PeriodEnd:   "2033-12-31",
	})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	handler := &api.CreateGoalHandler{Service: application.NewService(ctx, db)}
	authCtx := performanceAuthContext(userID, businessID)
	_, err = handler.Serve(authCtx, &api.CreateGoalRequest{
		EmployeeID: employeeID, ReviewCycleID: cycle.ID, Title: "Invalid weight", Weight: -1,
	})
	if status := performanceHTTPStatus(err); status != http.StatusBadRequest {
		t.Fatalf("invalid weight status = %d, want 400 (err=%v)", status, err)
	}
	created, err := handler.Serve(authCtx, &api.CreateGoalRequest{
		EmployeeID: employeeID, ReviewCycleID: cycle.ID, Title: "Own goal", Weight: 10,
	})
	if err != nil || created.Goal == nil {
		t.Fatalf("own goal create: response=%+v err=%v", created, err)
	}
	_, err = handler.Serve(authCtx, &api.CreateGoalRequest{
		EmployeeID: peerEmployeeID, ReviewCycleID: cycle.ID, Title: "Peer goal", Weight: 10,
	})
	if status := performanceHTTPStatus(err); status != http.StatusForbidden {
		t.Fatalf("peer goal status = %d, want 403 (err=%v)", status, err)
	}
	_, err = handler.Serve(authCtx, &api.CreateGoalRequest{
		EmployeeID: foreignEmployeeID, ReviewCycleID: cycle.ID, Title: "Foreign goal", Weight: 10,
	})
	if status := performanceHTTPStatus(err); status != http.StatusNotFound {
		t.Fatalf("foreign goal status = %d, want 404 (err=%v)", status, err)
	}
}

func TestGeneratedPerformanceListsPageAndPreserveOptionalFilters(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := performanceAPITestID(9_400_000_000)
	userID := performanceAPITestID(9_500_000_000)
	employeeID := performanceAPITestID(9_600_000_000)
	otherEmployeeID := employeeID + 1
	ensurePerformanceBusiness(t, db, businessID)
	grantPerformanceBusiness(t, db, userID, businessID)
	insertPerformanceAPIMployee(t, db, businessID, employeeID, &userID)
	insertPerformanceAPIMployee(t, db, businessID, otherEmployeeID, nil)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_goals WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_performance_review_cycles WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?", userID, businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id IN (?, ?)", employeeID, otherEmployeeID)
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", businessID)
	})

	repo := infrastructure.NewRepository(db)
	cycle, err := repo.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
		Name:        fmt.Sprintf("Generated list cycle %d", businessID),
		CycleType:   domain.CycleTypeAnnual,
		PeriodStart: "2034-01-01",
		PeriodEnd:   "2034-12-31",
	})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	for i := 1; i < 21; i++ {
		if _, err := repo.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
			Name:        fmt.Sprintf("Generated list cycle %d-%d", businessID, i),
			CycleType:   domain.CycleTypeAnnual,
			PeriodStart: "2034-01-01",
			PeriodEnd:   "2034-12-31",
		}); err != nil {
			t.Fatalf("create cycle %d: %v", i, err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err := repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
			EmployeeID: employeeID, ReviewCycleID: cycle.ID, Title: fmt.Sprintf("Generated goal %d", i), Weight: 1,
		}); err != nil {
			t.Fatalf("create goal: %v", err)
		}
	}
	if _, err := repo.CreateGoal(ctx, businessID, domain.CreateGoalParams{
		EmployeeID: otherEmployeeID, ReviewCycleID: cycle.ID, Title: "Other generated goal", Weight: 1,
	}); err != nil {
		t.Fatalf("create other goal: %v", err)
	}

	mux := http.NewServeMux()
	hrperformance.NewModule(ctx, db).RegisterAPI(mux)
	authCtx := performanceAuthContext(userID, businessID)

	request := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(authCtx)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, body=%s", path, rec.Code, rec.Body.String())
		}
		return rec
	}

	cyclePage := request("/hr/performance/v1/review-cycles?limit=1&offset=20")
	var cycles hrperformancev1.ListReviewCyclesResponse
	if err := json.NewDecoder(bytes.NewReader(cyclePage.Body.Bytes())).Decode(&cycles); err != nil {
		t.Fatalf("decode cycle page: %v", err)
	}
	if cycles.GetTotal() != 21 || cycles.GetLimit() != 1 || cycles.GetOffset() != 20 || len(cycles.GetCycles()) != 1 {
		t.Fatalf("cycle page metadata = total %d limit %d offset %d rows %d", cycles.GetTotal(), cycles.GetLimit(), cycles.GetOffset(), len(cycles.GetCycles()))
	}

	goalPage := request(fmt.Sprintf("/hr/performance/v1/goals?employee_id=%d&limit=1&offset=1", employeeID))
	var goals hrperformancev1.ListGoalsResponse
	if err := json.NewDecoder(bytes.NewReader(goalPage.Body.Bytes())).Decode(&goals); err != nil {
		t.Fatalf("decode goal page: %v", err)
	}
	if goals.GetTotal() != 2 || goals.GetLimit() != 1 || goals.GetOffset() != 1 || len(goals.GetGoals()) != 1 {
		t.Fatalf("goal page metadata = total %d limit %d offset %d rows %d", goals.GetTotal(), goals.GetLimit(), goals.GetOffset(), len(goals.GetGoals()))
	}
	for _, goal := range goals.GetGoals() {
		if goal.GetEmployeeId() != employeeID {
			t.Fatalf("employee filter leaked employee %d", goal.GetEmployeeId())
		}
	}
}
