package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

type breakTestRepo struct {
	mockRepo
	items       []domain.BreakRecord
	openByEmp   *domain.BreakRecord
	openByStaff *domain.BreakRecord
	byID        map[int64]*domain.BreakRecord
	lastList    domain.ListBreaksParams
	lastCreated domain.StartBreakInput
	endedID     int64
}

func (r *breakTestRepo) ListBreaks(_ context.Context, _ int64, params domain.ListBreaksParams) ([]domain.BreakRecord, error) {
	r.lastList = params
	return r.items, nil
}

func (r *breakTestRepo) GetBreak(_ context.Context, _, id int64) (*domain.BreakRecord, error) {
	if b, ok := r.byID[id]; ok {
		return b, nil
	}
	return nil, domain.ErrBreakNotFound
}

func (r *breakTestRepo) GetOpenBreakByEmployee(_ context.Context, _, _ int64) (*domain.BreakRecord, error) {
	return r.openByEmp, nil
}

func (r *breakTestRepo) GetOpenBreakByStaff(_ context.Context, _, _ int64) (*domain.BreakRecord, error) {
	return r.openByStaff, nil
}

func (r *breakTestRepo) CreateBreak(_ context.Context, businessID int64, input domain.StartBreakInput) (*domain.BreakRecord, error) {
	r.lastCreated = input
	return &domain.BreakRecord{
		ID:         77,
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: input.EmployeeID,
		StaffID:    input.StaffID,
		StaffName:  input.StaffName,
		BreakType:  input.BreakType,
		StartedAt:  time.Now().UTC(),
	}, nil
}

func (r *breakTestRepo) EndBreak(_ context.Context, businessID, id int64) (*domain.BreakRecord, error) {
	r.endedID = id
	now := time.Now().UTC()
	dur := int32(15)
	base := r.byID[id]
	if base == nil {
		return nil, domain.ErrBreakNotFound
	}
	cp := *base
	cp.BusinessID = businessID
	cp.EndedAt = &now
	cp.DurationMinutes = &dur
	return &cp, nil
}

type breakTestPermissions struct {
	mockPermissions
	names map[int64]string
}

func (p *breakTestPermissions) GetEmployeeDisplayName(_ context.Context, _, employeeID int64) (string, error) {
	return p.names[employeeID], nil
}

func TestStartBreak_EmployeeSelfServiceAndCrossEmployeeDenied(t *testing.T) {
	repo := &breakTestRepo{}
	perms := &breakTestPermissions{
		mockPermissions: mockPermissions{employeeID: 42},
		names:           map[int64]string{42: "Nadia Barista"},
	}
	svc := NewService(context.Background(), repo, nil, perms, &mockStaffResolver{staffByUser: 101})

	rec, err := svc.StartBreak(context.Background(), 10, 31, false, domain.StartBreakInput{BreakType: "Lunch"})
	if err != nil {
		t.Fatalf("StartBreak self: %v", err)
	}
	if rec.EmployeeID != 42 || rec.StaffID != 101 || rec.StaffName != "Nadia Barista" || rec.BreakType != "Lunch" {
		t.Fatalf("unexpected break record: %+v", rec)
	}

	_, err = svc.StartBreak(context.Background(), 10, 31, false, domain.StartBreakInput{EmployeeID: 99, BreakType: "Lunch"})
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied for another employee, got %v", err)
	}
}

func TestStartBreak_OwnerWithoutEmployeeRowSucceeds(t *testing.T) {
	repo := &breakTestRepo{}
	svc := NewService(context.Background(), repo, nil, &mockPermissions{}, &mockStaffResolver{userErr: domain.ErrStaffNotLinked})

	rec, err := svc.StartBreak(context.Background(), 10, 29, true, domain.StartBreakInput{BreakType: ""})
	if err != nil {
		t.Fatalf("StartBreak owner without employee row: %v", err)
	}
	if rec.StaffID != 29 || rec.BreakType != domain.BreakTypeLunch {
		t.Fatalf("unexpected owner break record: %+v", rec)
	}
}

func TestStartBreak_RejectsDuplicateOpenBreak(t *testing.T) {
	repo := &breakTestRepo{
		openByEmp: &domain.BreakRecord{ID: 5, EmployeeID: 42},
	}
	perms := &breakTestPermissions{mockPermissions: mockPermissions{employeeID: 42}}
	svc := NewService(context.Background(), repo, nil, perms, &mockStaffResolver{})

	_, err := svc.StartBreak(context.Background(), 10, 31, false, domain.StartBreakInput{BreakType: "Short"})
	if !errors.Is(err, domain.ErrBreakAlreadyOpen) {
		t.Fatalf("expected ErrBreakAlreadyOpen, got %v", err)
	}
}

func TestListBreaks_ScopesOrdinaryEmployeeToSelf(t *testing.T) {
	repo := &breakTestRepo{}
	perms := &breakTestPermissions{mockPermissions: mockPermissions{employeeID: 42}}
	svc := NewService(context.Background(), repo, nil, perms, &mockStaffResolver{})

	if _, err := svc.ListBreaks(context.Background(), 10, 31, false, 0, 0); err != nil {
		t.Fatalf("ListBreaks employee: %v", err)
	}
	if repo.lastList.EmployeeID == nil || *repo.lastList.EmployeeID != 42 {
		t.Fatalf("expected employee filter 42, got %+v", repo.lastList.EmployeeID)
	}

	if _, err := svc.ListBreaks(context.Background(), 10, 29, true, 0, 0); err != nil {
		t.Fatalf("ListBreaks manager: %v", err)
	}
	if repo.lastList.EmployeeID != nil {
		t.Fatalf("expected unfiltered list for manager, got %+v", *repo.lastList.EmployeeID)
	}
}

func TestEndBreak_EnforcesOwnershipForNonManager(t *testing.T) {
	repo := &breakTestRepo{
		byID: map[int64]*domain.BreakRecord{
			10: {ID: 10, EmployeeID: 42, BreakType: "Lunch"},
			11: {ID: 11, EmployeeID: 99, BreakType: "Lunch"},
		},
	}
	perms := &breakTestPermissions{mockPermissions: mockPermissions{employeeID: 42}}
	svc := NewService(context.Background(), repo, nil, perms, &mockStaffResolver{})

	if _, err := svc.EndBreak(context.Background(), 10, 31, 11, false); !errors.Is(err, domain.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied ending another employee's break, got %v", err)
	}

	ended, err := svc.EndBreak(context.Background(), 10, 31, 10, false)
	if err != nil {
		t.Fatalf("EndBreak own break: %v", err)
	}
	if ended.EndedAt == nil || ended.DurationMinutes == nil {
		t.Fatalf("expected ended break with duration, got %+v", ended)
	}
}
