package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

type w2PermissionProvider struct {
	employeeByUser     map[int64]int64
	businessByEmployee map[int64]int64
	permissions        map[int64]map[string]bool
	roles              map[int64]map[int64]bool
}

type invoiceExistsFunc func(context.Context, int64, int64) (bool, error)

func (f invoiceExistsFunc) InvoiceExists(ctx context.Context, businessID, invoiceID int64) (bool, error) {
	return f(ctx, businessID, invoiceID)
}

func newW2PermissionProvider() *w2PermissionProvider {
	return &w2PermissionProvider{
		employeeByUser:     make(map[int64]int64),
		businessByEmployee: make(map[int64]int64),
		permissions:        make(map[int64]map[string]bool),
		roles:              make(map[int64]map[int64]bool),
	}
}

func (p *w2PermissionProvider) GetEmployeeIDByUserID(_ context.Context, _, userID int64) (int64, error) {
	employeeID, ok := p.employeeByUser[userID]
	if !ok {
		return 0, domain.ErrTimeEntryAccessDenied
	}
	return employeeID, nil
}

func (p *w2PermissionProvider) GetEmployeeBusinessID(_ context.Context, _, employeeID int64) (int64, error) {
	businessID, ok := p.businessByEmployee[employeeID]
	if !ok {
		return 0, domain.ErrTimeEntryAccessDenied
	}
	return businessID, nil
}

func (p *w2PermissionProvider) HasPermission(_ context.Context, _, employeeID int64, permissionKey string) (bool, error) {
	return p.permissions[employeeID][permissionKey], nil
}

func (p *w2PermissionProvider) HasRole(_ context.Context, _, employeeID, roleID int64) (bool, error) {
	return p.roles[employeeID][roleID], nil
}

func (p *w2PermissionProvider) setPermission(employeeID int64, permission string, allowed bool) {
	if p.permissions[employeeID] == nil {
		p.permissions[employeeID] = make(map[string]bool)
	}
	p.permissions[employeeID][permission] = allowed
}

func (p *w2PermissionProvider) setRole(employeeID, roleID int64, allowed bool) {
	if p.roles[employeeID] == nil {
		p.roles[employeeID] = make(map[int64]bool)
	}
	p.roles[employeeID][roleID] = allowed
}

func newAllowApprovalPermissions(userEmployees map[int64]int64) *w2PermissionProvider {
	permissions := newW2PermissionProvider()
	for userID, employeeID := range userEmployees {
		permissions.employeeByUser[userID] = employeeID
		permissions.setPermission(employeeID, domain.PermissionApproveTimeOff, true)
		permissions.setRole(employeeID, 1, true)
		permissions.setRole(employeeID, 2, true)
	}
	return permissions
}

type w2ApprovalChain struct {
	levels []domain.ApprovalLevelInfo
}

func (c w2ApprovalChain) GetApprovalChain(context.Context, int64, string) (*domain.ApprovalChainInfo, error) {
	levels := append([]domain.ApprovalLevelInfo(nil), c.levels...)
	return &domain.ApprovalChainInfo{TotalLevels: int32(len(levels)), Levels: levels}, nil // #nosec G115 -- test fixtures contain only small explicit level counts.
}

func w2ptr[T any](value T) *T { return &value }

func TestCreateTimeEntry_RequiresOwnershipOrManageEmployeesPermission(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110001
	const callerUserID int64 = 7101
	const callerEmployeeID int64 = 5101
	const targetEmployeeID int64 = 5102

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[callerUserID] = callerEmployeeID
	permissions.businessByEmployee[callerEmployeeID] = businessID
	permissions.businessByEmployee[targetEmployeeID] = businessID
	service := NewService(ctx, infrastructure.NewRepository(db), nil, permissions)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
	})

	_, err := service.CreateTimeEntry(ctx, businessID, callerUserID, domain.TimeEntryRecord{
		EmployeeID:      targetEmployeeID,
		Date:            "2026-11-02",
		DurationMinutes: 60,
	})
	if !errors.Is(err, domain.ErrTimeEntryAccessDenied) {
		t.Fatalf("cross-employee create error = %v, want ErrTimeEntryAccessDenied", err)
	}

	permissions.setPermission(callerEmployeeID, domain.PermissionManageEmployees, true)
	created, err := service.CreateTimeEntry(ctx, businessID, callerUserID, domain.TimeEntryRecord{
		EmployeeID:      targetEmployeeID,
		Date:            "2026-11-02",
		DurationMinutes: 60,
	})
	if err != nil {
		t.Fatalf("manage-employees create: %v", err)
	}
	if created.EmployeeID != targetEmployeeID {
		t.Fatalf("created employee = %d, want %d", created.EmployeeID, targetEmployeeID)
	}
}

func TestDeleteTimeEntry_RejectsAnotherEmployeesEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110002
	const callerUserID int64 = 7201
	const callerEmployeeID int64 = 5201
	const ownerEmployeeID int64 = 5202

	repo := infrastructure.NewRepository(db)
	entry, err := repo.CreateTimeEntry(ctx, businessID, domain.TimeEntryRecord{
		EmployeeID:      ownerEmployeeID,
		Date:            "2026-11-03",
		DurationMinutes: 30,
	})
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
	})

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[callerUserID] = callerEmployeeID
	permissions.businessByEmployee[callerEmployeeID] = businessID
	permissions.businessByEmployee[ownerEmployeeID] = businessID
	service := NewService(ctx, infrastructure.NewRepository(db), nil, permissions)

	if err := service.DeleteTimeEntry(ctx, businessID, callerUserID, entry.ID); !errors.Is(err, domain.ErrTimeEntryAccessDenied) {
		t.Fatalf("unauthorized delete error = %v, want ErrTimeEntryAccessDenied", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM hr_time_entries WHERE business_id = ? AND id = ?", businessID, entry.ID).Scan(&count); err != nil {
		t.Fatalf("count entry after refusal: %v", err)
	}
	if count != 1 {
		t.Fatalf("entry count after unauthorized delete = %d, want 1", count)
	}
}

func TestGetTimesheetForCaller_RejectsAnotherEmployeesTimesheet(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110003
	const callerUserID int64 = 7301
	const callerEmployeeID int64 = 5301
	const ownerEmployeeID int64 = 5302

	repo := infrastructure.NewRepository(db)
	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  ownerEmployeeID,
		PeriodStart: "2026-11-01",
		PeriodEnd:   "2026-11-30",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("seed timesheet: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheets WHERE business_id = ?", businessID)
	})

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[callerUserID] = callerEmployeeID
	permissions.businessByEmployee[callerEmployeeID] = businessID
	service := NewService(ctx, infrastructure.NewRepository(db), nil, permissions)

	if _, err := service.GetTimesheetForCaller(ctx, businessID, callerUserID, ts.ID); !errors.Is(err, domain.ErrTimesheetAccessDenied) {
		t.Fatalf("other-employee detail error = %v, want ErrTimesheetAccessDenied", err)
	}
	permissions.setPermission(callerEmployeeID, domain.PermissionApproveTimeOff, true)
	if _, err := service.GetTimesheetForCaller(ctx, businessID, callerUserID, ts.ID); err != nil {
		t.Fatalf("approver detail read: %v", err)
	}
}

func TestApproveTimesheet_ChecksPermissionDesignatedApproverAndSelfApproval(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110004
	const ownerEmployeeID int64 = 5401
	const designatedEmployeeID int64 = 5402
	const noPermissionUserID int64 = 7401
	const wrongApproverUserID int64 = 7402
	const selfApproverUserID int64 = 7403
	const designatedApproverUserID int64 = 7404

	repo := infrastructure.NewRepository(db)
	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:           ownerEmployeeID,
		PeriodStart:          "2026-11-01",
		PeriodEnd:            "2026-11-30",
		Status:               domain.TimesheetStatusSubmitted,
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("seed submitted timesheet: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheets WHERE business_id = ?", businessID)
	})

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[noPermissionUserID] = 5403
	permissions.employeeByUser[wrongApproverUserID] = 5404
	permissions.employeeByUser[selfApproverUserID] = ownerEmployeeID
	permissions.employeeByUser[designatedApproverUserID] = designatedEmployeeID
	permissions.setPermission(5404, domain.PermissionApproveTimeOff, true)
	permissions.setPermission(ownerEmployeeID, domain.PermissionApproveTimeOff, true)
	permissions.setPermission(designatedEmployeeID, domain.PermissionApproveTimeOff, true)
	chain := w2ApprovalChain{levels: []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: designatedEmployeeID}}}
	service := NewService(ctx, infrastructure.NewRepository(db), chain, permissions)

	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, noPermissionUserID); !errors.Is(err, domain.ErrTimesheetApproverNotAllowed) {
		t.Fatalf("no-permission approval error = %v, want ErrTimesheetApproverNotAllowed", err)
	}
	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, wrongApproverUserID); !errors.Is(err, domain.ErrTimesheetApproverNotAllowed) {
		t.Fatalf("wrong-approver error = %v, want ErrTimesheetApproverNotAllowed", err)
	}
	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, selfApproverUserID); !errors.Is(err, domain.ErrTimesheetSelfApproval) {
		t.Fatalf("self-approval error = %v, want ErrTimesheetSelfApproval", err)
	}
	approved, err := service.ApproveTimesheet(ctx, businessID, ts.ID, designatedApproverUserID)
	if err != nil {
		t.Fatalf("designated approval: %v", err)
	}
	if approved.Status != domain.TimesheetStatusLocked {
		t.Fatalf("approved status = %q, want %q", approved.Status, domain.TimesheetStatusLocked)
	}
}

func TestUpdateTimeEntry_PreservesOmittedWorkflowFields(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110005
	const callerUserID int64 = 7501
	const employeeID int64 = 5501

	repo := infrastructure.NewRepository(db)
	ts, err := repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  employeeID,
		PeriodStart: "2026-11-01",
		PeriodEnd:   "2026-11-30",
		Status:      domain.TimesheetStatusDraft,
	})
	if err != nil {
		t.Fatalf("seed timesheet: %v", err)
	}
	permissions := newW2PermissionProvider()
	permissions.employeeByUser[callerUserID] = employeeID
	permissions.businessByEmployee[employeeID] = businessID
	service := NewService(ctx, infrastructure.NewRepository(db), nil, permissions)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheets WHERE business_id = ?", businessID)
	})

	created, err := service.CreateTimeEntry(ctx, businessID, callerUserID, domain.TimeEntryRecord{
		EmployeeID:      employeeID,
		TimesheetID:     ts.ID,
		ProjectID:       77,
		Date:            "2026-11-04",
		DurationMinutes: 60,
		Description:     "before",
		IsBillable:      true,
		Status:          domain.EntryStatusRejected,
		RejectionReason: "needs detail",
		ApprovedBy:      999,
		Source:          domain.EntrySourceImport,
	})
	if err != nil {
		t.Fatalf("seed entry: %v", err)
	}

	updated, err := service.UpdateTimeEntry(ctx, businessID, callerUserID, created.ID, domain.TimeEntryPatch{
		DurationMinutes: w2ptr(int32(90)),
		Description:     w2ptr("after"),
		IsBillable:      w2ptr(false),
	})
	if err != nil {
		t.Fatalf("update entry: %v", err)
	}
	if updated.EmployeeID != employeeID || updated.TimesheetID != ts.ID || updated.ProjectID != 77 {
		t.Fatalf("ownership/link fields changed: %+v", updated)
	}
	if updated.Status != domain.EntryStatusRejected || updated.RejectionReason != "needs detail" || updated.ApprovedBy != 999 {
		t.Fatalf("workflow fields changed: status=%q rejection=%q approved_by=%d", updated.Status, updated.RejectionReason, updated.ApprovedBy)
	}
	if updated.Source != domain.EntrySourceImport || updated.Description != "after" || updated.DurationMinutes != 90 || updated.IsBillable {
		t.Fatalf("patched fields not applied correctly: %+v", updated)
	}
}

func TestFinalTimesheetApprovalMarksEntriesApprovedBeforeInvoicing(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9110006
	const otherBusinessID int64 = 9111006
	const employeeUserID int64 = 7601
	const employeeID int64 = 5601
	const approverUserID int64 = 7602
	const approverEmployeeID int64 = 5602

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[employeeUserID] = employeeID
	permissions.employeeByUser[approverUserID] = approverEmployeeID
	permissions.setPermission(approverEmployeeID, domain.PermissionApproveTimeOff, true)
	chain := w2ApprovalChain{levels: []domain.ApprovalLevelInfo{{LevelOrder: 1, ApproverEmployeeID: approverEmployeeID}}}
	service := NewService(ctx, infrastructure.NewRepository(db), chain, permissions)
	service.SetInvoiceExistenceProvider(invoiceExistsFunc(func(ctx context.Context, businessID, invoiceID int64) (bool, error) {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM finance_invoices WHERE business_id = ? AND id = ?)", businessID, invoiceID).Scan(&exists)
		return exists, err
	}))

	ts, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-11-01", "2026-11-30")
	if err != nil {
		t.Fatalf("create timesheet: %v", err)
	}
	entry, err := service.CreateTimeEntry(ctx, businessID, employeeUserID, domain.TimeEntryRecord{
		EmployeeID:      employeeID,
		Date:            "2026-11-05",
		DurationMinutes: 120,
		Description:     "billable work",
		IsBillable:      true,
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := service.SubmitTimesheet(ctx, businessID, ts.ID); err != nil {
		t.Fatalf("submit timesheet: %v", err)
	}
	if _, err := service.ApproveTimesheet(ctx, businessID, ts.ID, approverUserID); err != nil {
		t.Fatalf("approve timesheet: %v", err)
	}

	repo := infrastructure.NewRepository(db)
	stored, err := repo.GetTimeEntry(ctx, businessID, entry.ID)
	if err != nil {
		t.Fatalf("read approved entry: %v", err)
	}
	if stored.Status != domain.EntryStatusApproved || stored.ApprovedBy != approverUserID || stored.TimesheetID != ts.ID {
		t.Fatalf("entry workflow after approval = status %q approved_by %d timesheet %d", stored.Status, stored.ApprovedBy, stored.TimesheetID)
	}

	insertInvoice := func(invoiceBusinessID int64) int64 {
		result, err := db.ExecContext(ctx, `
			INSERT INTO finance_invoices
				(owner_id, business_id, invoice_number, client_name, status, issue_date, due_date)
			VALUES (?, ?, ?, 'W2 Test Client', 'sent', '2026-11-01', '2026-12-01')`,
			invoiceBusinessID, invoiceBusinessID, fmt.Sprintf("W2-%d", invoiceBusinessID))
		if err != nil {
			t.Fatalf("insert invoice for business %d: %v", invoiceBusinessID, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("invoice id: %v", err)
		}
		return id
	}
	otherInvoiceID := insertInvoice(otherBusinessID)
	invoiceID := insertInvoice(businessID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_time_entries WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_timesheets WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id IN (?, ?)", businessID, otherBusinessID)
	})

	if _, err := service.MarkEntriesInvoiced(ctx, businessID, otherInvoiceID, []int64{entry.ID}); !errors.Is(err, domain.ErrInvoiceNotFound) {
		t.Fatalf("foreign invoice error = %v, want ErrInvoiceNotFound", err)
	}
	if _, err := service.MarkEntriesInvoiced(ctx, businessID, 9_999_999, []int64{entry.ID}); !errors.Is(err, domain.ErrInvoiceNotFound) {
		t.Fatalf("missing invoice error = %v, want ErrInvoiceNotFound", err)
	}
	stored, err = repo.GetTimeEntry(ctx, businessID, entry.ID)
	if err != nil {
		t.Fatalf("read entry after foreign invoice: %v", err)
	}
	if stored.BillableInvoiceID != 0 {
		t.Fatalf("foreign invoice stamped entry with invoice %d", stored.BillableInvoiceID)
	}

	marked, err := service.MarkEntriesInvoiced(ctx, businessID, invoiceID, []int64{entry.ID})
	if err != nil {
		t.Fatalf("mark valid invoice: %v", err)
	}
	if marked.MarkedCount != 1 || marked.SkippedCount != 0 {
		t.Fatalf("mark result = %+v, want one marked and none skipped", marked)
	}
	stored, err = repo.GetTimeEntry(ctx, businessID, entry.ID)
	if err != nil {
		t.Fatalf("read invoiced entry: %v", err)
	}
	if stored.BillableInvoiceID != invoiceID {
		t.Fatalf("billable_invoice_id = %d, want %d", stored.BillableInvoiceID, invoiceID)
	}
}
