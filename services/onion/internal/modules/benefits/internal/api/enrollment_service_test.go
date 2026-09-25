package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/benefits/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	hrbenefitv1 "github.com/voxire/voxie/gen/go/hr/benefit/v1"
)

// apiBusinessID keeps these API-level rows out of the way of the repository
// tests, which work in testutil.OwnerID's scope.
const apiBusinessID int64 = 5100

const (
	userAlice int64 = 5201
	userBob   int64 = 5202

	employeeAlice int64 = 5301
	employeeBob   int64 = 5302
)

// fakeDirectory stands in for hr_core's user -> employee lookup.
type fakeDirectory struct {
	byUser map[int64]int64
}

func (f fakeDirectory) GetEmployeeIDByUserID(_ context.Context, _, userID int64) (int64, error) {
	if id, ok := f.byUser[userID]; ok {
		return id, nil
	}
	return 0, errors.New("no employee record for user")
}

func newTestAPI(t *testing.T) (*BenefitAPI, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)

	directory := fakeDirectory{byUser: map[int64]int64{
		userAlice: employeeAlice,
		userBob:   employeeBob,
	}}
	service := application.NewService(context.Background(), db, directory)

	t.Cleanup(func() {
		for _, employeeID := range []int64{employeeAlice, employeeBob} {
			if _, err := db.ExecContext(context.Background(),
				"DELETE FROM hr_benefit_enrollments WHERE business_id = ? AND employee_id = ?",
				apiBusinessID, employeeID); err != nil {
				t.Logf("cleanup enrollments for employee %d: %v", employeeID, err)
			}
		}
		if _, err := db.ExecContext(context.Background(),
			"DELETE FROM hr_benefit_plans WHERE business_id = ?", apiBusinessID); err != nil {
			t.Logf("cleanup plans: %v", err)
		}
	})

	return NewBenefitAPI(service), db
}

// ctxFor builds a request context authenticated as the given platform user.
func ctxFor(userID int64) context.Context {
	device := authPkg.Device{
		User: &authPkg.AuthUser{
			UserID:     userID,
			BusinessID: apiBusinessID,
		},
		Language: "en",
	}
	return device.Ctx(context.Background())
}

func newPlanViaAPI(t *testing.T, api *BenefitAPI, userID int64, name string) int64 {
	t.Helper()
	res, err := api.CreateBenefit(ctxFor(userID), &hrbenefitv1.CreateBenefitRequest{
		Name: name, Type: "health", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateBenefit %q: %v", name, err)
	}
	return res.GetId()
}

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) {
		t.Fatalf("error %v is not an onekit http error", err)
	}
	return httpErr.HTTPStatusCode()
}

func TestEnrollInBenefit_ReturnsEnrollmentForCallingEmployee(t *testing.T) {
	api, _ := newTestAPI(t)
	planID := newPlanViaAPI(t, api, userAlice, "API Enroll Plan")

	res, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)})
	if err != nil {
		t.Fatalf("EnrollInBenefit: %v", err)
	}

	got := res.GetEnrollment()
	if got == nil {
		t.Fatal("expected an enrollment in the response")
	}
	if got.GetEmployeeId() != employeeAlice {
		t.Errorf("EmployeeId = %d, want %d resolved from auth context", got.GetEmployeeId(), employeeAlice)
	}
	if got.GetPlanId() != planID {
		t.Errorf("PlanId = %d, want %d", got.GetPlanId(), planID)
	}
	if got.GetStatus() != "active" {
		t.Errorf("Status = %q, want active", got.GetStatus())
	}
	if got.GetEnrolledAt() == "" {
		t.Error("EnrolledAt is empty, want an ISO-8601 timestamp")
	}
	if got.CancelledAt != nil {
		t.Errorf("CancelledAt = %v, want nil while active", got.CancelledAt)
	}
}

func TestListMyBenefitEnrollments_ReturnsOwnEnrollment(t *testing.T) {
	api, _ := newTestAPI(t)
	planID := newPlanViaAPI(t, api, userAlice, "API List Plan")

	if _, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)}); err != nil {
		t.Fatalf("EnrollInBenefit: %v", err)
	}

	res, err := api.ListMyBenefitEnrollments(ctxFor(userAlice), &hrbenefitv1.ListMyBenefitEnrollmentsRequest{})
	if err != nil {
		t.Fatalf("ListMyBenefitEnrollments: %v", err)
	}
	if len(res.GetEnrollments()) != 1 {
		t.Fatalf("got %d enrollments, want 1", len(res.GetEnrollments()))
	}
	if res.GetEnrollments()[0].GetPlanId() != planID {
		t.Errorf("PlanId = %d, want %d", res.GetEnrollments()[0].GetPlanId(), planID)
	}
	if res.GetPagination().GetTotal() != 1 {
		t.Errorf("pagination.total = %d, want 1", res.GetPagination().GetTotal())
	}
}

// The scoping guarantee: Bob must never see Alice's enrollments.
func TestListMyBenefitEnrollments_SecondEmployeeSeesNone(t *testing.T) {
	api, _ := newTestAPI(t)
	planID := newPlanViaAPI(t, api, userAlice, "API Scoping Plan")

	if _, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)}); err != nil {
		t.Fatalf("EnrollInBenefit as alice: %v", err)
	}

	res, err := api.ListMyBenefitEnrollments(ctxFor(userBob), &hrbenefitv1.ListMyBenefitEnrollmentsRequest{})
	if err != nil {
		t.Fatalf("ListMyBenefitEnrollments as bob: %v", err)
	}
	if len(res.GetEnrollments()) != 0 {
		t.Fatalf("bob sees %d of alice's enrollments, want 0", len(res.GetEnrollments()))
	}
	if res.GetPagination().GetTotal() != 0 {
		t.Errorf("pagination.total = %d, want 0 for bob", res.GetPagination().GetTotal())
	}
}

func TestEnrollInBenefit_DuplicateReturnsConflict(t *testing.T) {
	api, _ := newTestAPI(t)
	planID := newPlanViaAPI(t, api, userAlice, "API Duplicate Plan")

	if _, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)}); err != nil {
		t.Fatalf("first EnrollInBenefit: %v", err)
	}

	_, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)})
	if err == nil {
		t.Fatal("expected the duplicate enrollment to fail")
	}
	if got := statusOf(t, err); got != http.StatusConflict {
		t.Errorf("status = %d, want %d Conflict", got, http.StatusConflict)
	}
}

func TestEnrollInBenefit_UnknownPlanReturnsNotFound(t *testing.T) {
	api, _ := newTestAPI(t)

	_, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(999999999)})
	if err == nil {
		t.Fatal("expected enrolling in an unknown plan to fail")
	}
	if got := statusOf(t, err); got != http.StatusNotFound {
		t.Errorf("status = %d, want %d Not Found", got, http.StatusNotFound)
	}
}

func TestEnrollInBenefit_MissingPlanIDReturnsBadRequest(t *testing.T) {
	api, _ := newTestAPI(t)

	_, err := api.EnrollInBenefit(ctxFor(userAlice), &hrbenefitv1.EnrollInBenefitRequest{})
	if err == nil {
		t.Fatal("expected a missing plan_id to fail")
	}
	if got := statusOf(t, err); got != http.StatusBadRequest {
		t.Errorf("status = %d, want %d Bad Request", got, http.StatusBadRequest)
	}
}

func TestEnrollInBenefit_AnonymousIsDenied(t *testing.T) {
	api, _ := newTestAPI(t)

	_, err := api.EnrollInBenefit(context.Background(), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(1)})
	if err == nil {
		t.Fatal("expected an anonymous caller to be denied")
	}
	if got := statusOf(t, err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d Forbidden", got, http.StatusForbidden)
	}
}

func TestListMyBenefitEnrollments_AnonymousIsDenied(t *testing.T) {
	api, _ := newTestAPI(t)

	_, err := api.ListMyBenefitEnrollments(context.Background(), &hrbenefitv1.ListMyBenefitEnrollmentsRequest{})
	if err == nil {
		t.Fatal("expected an anonymous caller to be denied")
	}
	if got := statusOf(t, err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d Forbidden", got, http.StatusForbidden)
	}
}

// A user with no hr_employees row cannot enroll as somebody else.
func TestEnrollInBenefit_UserWithoutEmployeeRecordIsDenied(t *testing.T) {
	api, _ := newTestAPI(t)
	planID := newPlanViaAPI(t, api, userAlice, "API Unmapped Plan")

	const unmappedUser int64 = 5999
	_, err := api.EnrollInBenefit(ctxFor(unmappedUser), &hrbenefitv1.EnrollInBenefitRequest{PlanId: testutil.Int64Ptr(planID)})
	if err == nil {
		t.Fatal("expected a user with no employee record to be denied")
	}
	if got := statusOf(t, err); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d Forbidden", got, http.StatusForbidden)
	}
}

func TestListMyBenefitEnrollments_ReportsPageAndLimit(t *testing.T) {
	api, _ := newTestAPI(t)

	res, err := api.ListMyBenefitEnrollments(ctxFor(userAlice), &hrbenefitv1.ListMyBenefitEnrollmentsRequest{
		Page: 2, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListMyBenefitEnrollments: %v", err)
	}
	if res.GetPagination().GetPage() != 2 {
		t.Errorf("pagination.page = %d, want 2", res.GetPagination().GetPage())
	}
	if res.GetPagination().GetLimit() != 10 {
		t.Errorf("pagination.limit = %d, want 10", res.GetPagination().GetLimit())
	}
}
