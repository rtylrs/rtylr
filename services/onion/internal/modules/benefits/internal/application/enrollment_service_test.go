package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
)

type mockEnrollmentRepo struct {
	mockRepo

	created     domain.BenefitEnrollment
	createErr   error
	gotBusiness int64
	gotEmployee int64
	gotPlan     int64

	listItems  []domain.BenefitEnrollment
	listTotal  int32
	listErr    error
	gotLimit   int32
	gotOffset  int32
	gotListEmp int64
}

func (m *mockEnrollmentRepo) CreateEnrollment(_ context.Context, businessID, employeeID, planID int64) (domain.BenefitEnrollment, error) {
	m.gotBusiness = businessID
	m.gotEmployee = employeeID
	m.gotPlan = planID
	return m.created, m.createErr
}

func (m *mockEnrollmentRepo) ListEnrollmentsByEmployee(_ context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.BenefitEnrollment, int32, error) {
	m.gotBusiness = businessID
	m.gotListEmp = employeeID
	m.gotLimit = limit
	m.gotOffset = offset
	return m.listItems, m.listTotal, m.listErr
}

func TestEnroll_ForwardsScopeAndPlanToRepository(t *testing.T) {
	m := &mockEnrollmentRepo{created: domain.BenefitEnrollment{ID: 7, Status: "active"}}
	s := &Service{repo: m}

	got, err := s.Enroll(context.Background(), 4001, 55, 900)
	if err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
	if m.gotBusiness != 4001 || m.gotEmployee != 55 || m.gotPlan != 900 {
		t.Errorf("forwarded (business=%d employee=%d plan=%d), want (4001 55 900)",
			m.gotBusiness, m.gotEmployee, m.gotPlan)
	}
}

func TestEnroll_PropagatesAlreadyEnrolled(t *testing.T) {
	s := &Service{repo: &mockEnrollmentRepo{createErr: domain.ErrAlreadyEnrolled}}

	_, err := s.Enroll(context.Background(), 4002, 55, 900)
	if !errors.Is(err, domain.ErrAlreadyEnrolled) {
		t.Fatalf("error = %v, want ErrAlreadyEnrolled", err)
	}
}

func TestEnroll_PropagatesPlanNotFound(t *testing.T) {
	s := &Service{repo: &mockEnrollmentRepo{createErr: domain.ErrBenefitPlanNotFound}}

	_, err := s.Enroll(context.Background(), 4003, 55, 900)
	if !errors.Is(err, domain.ErrBenefitPlanNotFound) {
		t.Fatalf("error = %v, want ErrBenefitPlanNotFound", err)
	}
}

func TestListMyEnrollments_ReturnsItemsAndTotal(t *testing.T) {
	m := &mockEnrollmentRepo{
		listItems: []domain.BenefitEnrollment{{ID: 1}, {ID: 2}},
		listTotal: 9,
	}
	s := &Service{repo: m}

	items, total, err := s.ListMyEnrollments(context.Background(), 4004, 55, 1, 50)
	if err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
	if total != 9 {
		t.Errorf("total = %d, want 9 (the unpaginated count)", total)
	}
}

func TestListMyEnrollments_ScopesToRequestedEmployee(t *testing.T) {
	m := &mockEnrollmentRepo{}
	s := &Service{repo: m}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4005, 77, 1, 50); err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if m.gotListEmp != 77 {
		t.Errorf("employee_id forwarded = %d, want 77", m.gotListEmp)
	}
}

func TestListMyEnrollments_DefaultsLimitTo50(t *testing.T) {
	m := &mockEnrollmentRepo{}
	s := &Service{repo: m}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4006, 55, 1, 0); err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if m.gotLimit != 50 {
		t.Errorf("limit = %d, want the onekit default of 50", m.gotLimit)
	}
}

func TestListMyEnrollments_CapsLimitAt500(t *testing.T) {
	m := &mockEnrollmentRepo{}
	s := &Service{repo: m}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4007, 55, 1, 5000); err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if m.gotLimit != 500 {
		t.Errorf("limit = %d, want the onekit cap of 500", m.gotLimit)
	}
}

func TestListMyEnrollments_ComputesOffsetFromPage(t *testing.T) {
	m := &mockEnrollmentRepo{}
	s := &Service{repo: m}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4008, 55, 3, 20); err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if m.gotOffset != 40 {
		t.Errorf("offset = %d, want 40 for page 3 at limit 20", m.gotOffset)
	}
}

func TestListMyEnrollments_TreatsPageZeroAsFirstPage(t *testing.T) {
	m := &mockEnrollmentRepo{}
	s := &Service{repo: m}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4009, 55, 0, 20); err != nil {
		t.Fatalf("ListMyEnrollments: %v", err)
	}
	if m.gotOffset != 0 {
		t.Errorf("offset = %d, want 0 when page is unset", m.gotOffset)
	}
}

func TestListMyEnrollments_PropagatesError(t *testing.T) {
	want := errors.New("list enrollments failed")
	s := &Service{repo: &mockEnrollmentRepo{listErr: want}}

	if _, _, err := s.ListMyEnrollments(context.Background(), 4010, 55, 1, 50); !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
