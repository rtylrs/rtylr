package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/labourcost"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

const (
	hrPayrollScope = "hr_payroll"
	hrPayrollTTL   = 5 * time.Minute

	// payrollExportReleaseTimeout bounds the compensating release of an export
	// claim so a canceled request cannot leave the run stuck in 'exporting'.
	payrollExportReleaseTimeout = 5 * time.Second
)

func invalidateHRPayrollScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrPayrollScope, ownerID))
}

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
	permissions      domain.PermissionProvider
	financePoster    domain.FinanceJournalPoster
	notifier         domain.NotificationDispatcher
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider, permissions domain.PermissionProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
		permissions:      permissions,
		financePoster:    noopFinancePoster{},
		notifier:         noopNotifier{},
	}
}

func (s *Service) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return s.permissions.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

// GetProjectLabourCost implements the platform labour-cost port. The service
// deliberately exposes only HR's aggregate result, never hourly rates.
func (s *Service) GetProjectLabourCost(ctx context.Context, businessID, projectID int64) (labourcost.ProjectCost, error) {
	return s.repo.GetProjectLabourCost(ctx, businessID, projectID)
}

func (s *Service) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return s.permissions.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (s *Service) HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error) {
	return s.permissions.HasRole(ctx, ownerID, employeeID, roleID)
}

// SetFinancePoster wires a finance module implementation into the payroll
// service. Call this from the module bootstrap after both modules have been
// constructed. Until called, payroll exports fail instead of being marked as
// exported without a corresponding finance journal entry.
func (s *Service) SetFinancePoster(p domain.FinanceJournalPoster) {
	if p != nil {
		s.financePoster = p
	}
}

// SetNotifier wires a notification module implementation into the payroll
// service. Until called, notifications are no-ops (the warning is logged).
func (s *Service) SetNotifier(n domain.NotificationDispatcher) {
	if n != nil {
		s.notifier = n
	}
}

type noopFinancePoster struct{}

func (noopFinancePoster) PostPayrollJournal(ctx context.Context, entry domain.PayrollJournalEntry) error {
	logger.Ctx(ctx).Warn("finance journal poster not wired; payroll export skipped",
		logger.Any("owner_id", entry.OwnerID),
		logger.Any("run_id", entry.RunID),
	)
	return domain.ErrFinanceIntegrationUnavailable
}

type noopNotifier struct{}

func (noopNotifier) Dispatch(ctx context.Context, event domain.NotificationEvent) error {
	logger.Ctx(ctx).Warn("notification dispatcher not wired; event dropped",
		logger.Any("owner_id", event.OwnerID),
		logger.String("type", event.Type),
	)
	return nil
}

func (s *Service) GetPayrollConfig(ctx context.Context, ownerID int64) (*domain.PayrollConfigRecord, error) {
	return cacher.Of[*domain.PayrollConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrPayrollScope, ownerID, "config"),
		func(ctx context.Context) (*domain.PayrollConfigRecord, error) {
			cfg, err := s.repo.GetPayrollConfig(ctx, ownerID)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
		cacher.WithExpiry(ctx, hrPayrollTTL),
	)
}

func (s *Service) UpsertPayrollConfig(ctx context.Context, ownerID int64, cfg domain.PayrollConfigRecord) error {
	if cfg.PeriodType == "" {
		cfg.PeriodType = "monthly"
	}
	if cfg.PayDayOfMonth <= 0 || cfg.PayDayOfMonth > 28 {
		cfg.PayDayOfMonth = 28
	}
	defer invalidateHRPayrollScope(ctx, ownerID)
	return s.repo.UpsertPayrollConfig(ctx, ownerID, cfg)
}

func (s *Service) TriggerPayrollRun(ctx context.Context, ownerID, createdBy int64, periodStart, periodEnd string) (*domain.PayrollRunRecord, []domain.PayrollEntryRecord, error) {
	defer invalidateHRPayrollScope(ctx, ownerID)
	startDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, nil, domain.ErrInvalidDateRange
	}
	if endDate.Before(startDate) {
		return nil, nil, domain.ErrInvalidDateRange
	}

	existing, err := s.repo.GetPayrollRunByPeriod(ctx, ownerID, periodStart, periodEnd)
	if err != nil && !errors.Is(err, domain.ErrPayrollRunNotFound) {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, domain.ErrPayrollRunAlreadyExists
	}

	cfg, err := s.repo.GetPayrollConfig(ctx, ownerID)
	if err != nil {
		return nil, nil, err
	}
	if cfg == nil {
		return nil, nil, domain.ErrPayrollNotConfigured
	}

	timesheetSummaries, err := s.repo.ListApprovedTimesheetSummaries(ctx, ownerID, periodStart, periodEnd)
	if err != nil {
		return nil, nil, err
	}

	var overtimeSummaries []domain.OvertimeSummary
	var defaultMultiplier float64
	if cfg.OvertimeEnabled {
		mult, configured, otErr := s.repo.GetOvertimeConfig(ctx, ownerID)
		if otErr != nil {
			return nil, nil, otErr
		}
		if !configured {
			return nil, nil, domain.ErrOvertimeConfigMissing
		}
		defaultMultiplier = mult

		overtimeSummaries, err = s.repo.ListApprovedOvertimeSummaries(ctx, ownerID, periodStart, periodEnd)
		if err != nil {
			return nil, nil, err
		}
	}

	employees, err := s.repo.ListEmployeePayInfo(ctx, ownerID)
	if err != nil {
		return nil, nil, err
	}

	employeeMap := make(map[int64]domain.EmployeePayInfo, len(employees))
	for _, emp := range employees {
		employeeMap[emp.EmployeeID] = emp
	}

	tsMap := make(map[int64]int32, len(timesheetSummaries))
	for _, ts := range timesheetSummaries {
		tsMap[ts.EmployeeID] = ts.TotalMinutes
	}

	otMap := make(map[int64]domain.OvertimeSummary, len(overtimeSummaries))
	for _, ot := range overtimeSummaries {
		otMap[ot.EmployeeID] = ot
	}

	entries := make([]domain.PayrollEntryRecord, 0, len(employees))
	var totalGrossCents, totalNetCents int64

	for _, emp := range employees {
		hoursWorked := float64(tsMap[emp.EmployeeID]) / 60.0
		overtimeMinutes := int32(0)
		overtimeWeightedMinutes := float64(0)
		overtimeUnweightedMinutes := float64(0)

		if ot, ok := otMap[emp.EmployeeID]; ok {
			overtimeMinutes = ot.TotalMinutes
			overtimeWeightedMinutes = ot.WeightedMinutes
			overtimeUnweightedMinutes = ot.UnweightedMinutes
		}
		overtimeHours := float64(overtimeMinutes) / 60.0

		// The rate/salary fields are stored as integer cents; the proration
		// and multiplier arithmetic below is unchanged from before the
		// cents migration. Each computed pay component is rounded once
		// before the stored integer amounts are combined.
		hourlyRateDollars := money.CentsToDollars(emp.HourlyRateCents)
		monthlySalaryDollars := money.CentsToDollars(emp.MonthlySalaryCents)

		var basePay float64
		switch emp.PayType {
		case "hourly":
			// Overtime minutes are already inside hoursWorked; pay them at
			// the multiplier only, not once more inside the base rate.
			regularHours := hoursWorked - overtimeHours
			if regularHours < 0 {
				regularHours = 0
			}
			basePay = regularHours * hourlyRateDollars
		case "monthly":
			basePay = calculateMonthlyBasePay(monthlySalaryDollars, startDate, endDate)
		default:
			basePay = calculateMonthlyBasePay(monthlySalaryDollars, startDate, endDate)
		}

		overtimePay := 0.0
		if cfg.OvertimeEnabled && overtimeHours > 0 {
			if hourlyRateDollars <= 0 {
				return nil, nil, domain.ErrOvertimeRateMissing
			}
			// The repository returns minutes weighted by each request's own
			// multiplier. This avoids applying an unweighted average rate to
			// the aggregate overtime duration.
			overtimePay = calculateOvertimePay(hourlyRateDollars, overtimeWeightedMinutes, overtimeUnweightedMinutes, defaultMultiplier)
		}

		basePayCents, overtimePayCents, netPayCents := calculatePayCents(basePay, overtimePay)
		grossPayCents := netPayCents

		entry := domain.PayrollEntryRecord{
			OwnerID:          ownerID,
			EmployeeID:       emp.EmployeeID,
			HoursWorked:      hoursWorked,
			OvertimeHours:    overtimeHours,
			BasePayCents:     basePayCents,
			OvertimePayCents: overtimePayCents,
			BonusesCents:     0,
			DeductionsCents:  0,
			NetPayCents:      netPayCents,
		}

		entries = append(entries, entry)
		totalGrossCents += grossPayCents
		totalNetCents += netPayCents
	}

	run := domain.PayrollRunRecord{
		OwnerID:         ownerID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		Status:          "draft",
		CreatedBy:       createdBy,
		TotalGrossCents: totalGrossCents,
		TotalNetCents:   totalNetCents,
	}

	createdRun, createdEntries, err := s.repo.CreatePayrollRunWithEntries(ctx, ownerID, run, entries)
	if err != nil {
		return nil, nil, err
	}

	return createdRun, createdEntries, nil
}

func calculateOvertimePay(hourlyRateDollars, weightedMinutes, unweightedMinutes, defaultMultiplier float64) float64 {
	return (weightedMinutes + unweightedMinutes*defaultMultiplier) / 60.0 * hourlyRateDollars
}

func calculatePayCents(basePay, overtimePay float64) (basePayCents, overtimePayCents, netPayCents int64) {
	basePayCents = money.DollarsToCents(basePay)
	overtimePayCents = money.DollarsToCents(overtimePay)
	return basePayCents, overtimePayCents, basePayCents + overtimePayCents
}

// calculateMonthlyBasePay prorates a monthly salary over an inclusive calendar
// date range. Each calendar month contributes its salary fraction using that
// month's actual number of days, so a complete calendar month always pays the
// configured monthly salary regardless of whether it has 28, 29, 30, or 31
// days.
func calculateMonthlyBasePay(monthlySalary float64, startDate, endDate time.Time) float64 {
	if endDate.Before(startDate) {
		return 0
	}

	var basePay float64
	for current := startDate; !current.After(endDate); {
		monthEnd := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		segmentEnd := monthEnd
		if endDate.Before(segmentEnd) {
			segmentEnd = endDate
		}

		inclusiveDays := float64(int(segmentEnd.Sub(current)/(24*time.Hour)) + 1)
		basePay += monthlySalary * inclusiveDays / float64(monthEnd.Day())
		current = segmentEnd.AddDate(0, 0, 1)
	}

	return basePay
}

func (s *Service) ApprovePayrollRun(ctx context.Context, ownerID, runID, approverID int64) error {
	defer invalidateHRPayrollScope(ctx, ownerID)
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return err
	}
	if run.Status != domain.PayrollStatusDraft && run.Status != domain.PayrollStatusPendingApproval {
		return domain.ErrPayrollRunNotApprovable
	}
	if err := validatePayrollApprover(run, approverID); err != nil {
		return err
	}
	if err := s.repo.ValidatePayrollRunEmployees(ctx, ownerID, runID); err != nil {
		return err
	}

	chain, err := s.approvalProvider.GetApprovalChain(ctx, ownerID, "payroll")
	if err != nil {
		return err
	}
	if chain == nil || len(chain.Levels) == 0 || len(chain.Levels) != int(chain.TotalLevels) || len(chain.Levels) != 1 {
		if chain == nil {
			return domain.ErrApprovalChainNotConfigured
		}
		return domain.ErrPayrollApprovalChainInvalid
	}
	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, approverID)
	if err != nil {
		return err
	}
	allowed, err := s.approverMatchesChain(ctx, ownerID, employeeID, chain.Levels)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrPayrollApproverNotAllowed
	}

	if run.Status == domain.PayrollStatusDraft {
		if err := s.repo.UpdatePayrollRunStatus(ctx, ownerID, runID, domain.PayrollStatusPendingApproval); err != nil {
			return err
		}
	}

	return s.repo.SetPayrollRunApproved(ctx, ownerID, runID, approverID)
}

func validatePayrollApprover(run *domain.PayrollRunRecord, approverID int64) error {
	if run != nil && run.CreatedBy == approverID {
		return domain.ErrPayrollSelfApproval
	}
	return nil
}

func (s *Service) approverMatchesChain(ctx context.Context, ownerID, employeeID int64, levels []domain.ApprovalLevelInfo) (bool, error) {
	if len(levels) != 1 {
		return false, nil
	}
	level := levels[0]
	if level.LevelOrder > 0 && level.LevelOrder != 1 {
		return false, nil
	}
	if level.ApproverEmployeeID > 0 && level.ApproverEmployeeID == employeeID {
		return true, nil
	}
	if level.ApproverRoleID <= 0 {
		return false, nil
	}
	return s.permissions.HasRole(ctx, ownerID, employeeID, level.ApproverRoleID)
}

func (s *Service) ExportPayrollToFinance(ctx context.Context, ownerID, runID int64) error {
	defer invalidateHRPayrollScope(ctx, ownerID)
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return err
	}
	if err := s.repo.ValidatePayrollRunEmployees(ctx, ownerID, runID); err != nil {
		return err
	}

	if err := s.repo.ClaimPayrollRunForExport(ctx, ownerID, runID); err != nil {
		return err
	}
	if err := s.repo.ValidatePayrollRunEmployees(ctx, ownerID, runID); err != nil {
		s.releasePayrollExportClaim(ctx, ownerID, runID, "employee validation failed")
		return err
	}
	if s.financePoster == nil {
		s.releasePayrollExportClaim(ctx, ownerID, runID, "finance integration unavailable")
		return domain.ErrFinanceIntegrationUnavailable
	}

	// The finance journal contract (owned by the finance module, out of
	// scope for this fix) still carries these amounts as float64 dollars,
	// so the internally-stored cents are converted at this boundary and
	// never propagated further as a float.
	if postErr := s.financePoster.PostPayrollJournal(ctx, domain.PayrollJournalEntry{
		OwnerID:     ownerID,
		BusinessID:  run.BusinessID,
		RunID:       runID,
		PeriodStart: run.PeriodStart,
		PeriodEnd:   run.PeriodEnd,
		TotalGross:  money.CentsToDollars(run.TotalGrossCents),
		TotalNet:    money.CentsToDollars(run.TotalNetCents),
	}); postErr != nil {
		s.releasePayrollExportClaim(ctx, ownerID, runID, "finance journal post failed")
		logger.Ctx(ctx).Error("finance journal post failed",
			logger.Any("owner_id", ownerID),
			logger.Any("run_id", runID),
			logger.Error(postErr),
		)
		return postErr
	}

	if err := s.repo.SetPayrollRunExported(ctx, ownerID, runID); err != nil {
		logger.Ctx(ctx).Error("finance journal posted but payroll export could not be finalized",
			logger.Any("owner_id", ownerID), logger.Any("run_id", runID), logger.Error(err))
		return err
	}

	entries, listErr := s.repo.ListPayrollEntries(ctx, ownerID, runID)
	if listErr != nil {
		logger.Ctx(ctx).Warn("payroll exported but pay slip notifications could not be enumerated",
			logger.Any("owner_id", ownerID),
			logger.Any("run_id", runID),
			logger.Error(listErr),
		)
		return nil
	}

	for _, entry := range entries {
		if dispErr := s.notifier.Dispatch(ctx, domain.NotificationEvent{
			OwnerID: ownerID,
			Type:    "payroll.payslip",
			Subject: "Your pay slip is ready",
			Payload: map[string]any{
				"run_id":       runID,
				"employee_id":  entry.EmployeeID,
				"net_pay":      money.CentsToDollars(entry.NetPayCents),
				"period_start": run.PeriodStart,
				"period_end":   run.PeriodEnd,
			},
		}); dispErr != nil {
			logger.Ctx(ctx).Warn("pay slip notification failed",
				logger.Any("owner_id", ownerID),
				logger.Any("run_id", runID),
				logger.Any("employee_id", entry.EmployeeID),
				logger.Error(dispErr),
			)
		}
	}

	return nil
}

// releasePayrollExportClaim returns a run claimed for export to 'approved'.
// It deliberately detaches from the caller's context: releasing the claim is
// what makes a failed or abandoned export retryable, so it must still run
// after the request context is canceled or has hit its deadline, which is
// exactly when the client disconnected mid-export (B62-R05).
func (s *Service) releasePayrollExportClaim(ctx context.Context, ownerID, runID int64, reason string) {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), payrollExportReleaseTimeout)
	defer cancel()
	if err := s.repo.ReleasePayrollRunExportClaim(releaseCtx, ownerID, runID); err != nil {
		logger.Ctx(releaseCtx).Error("failed to release payroll export claim",
			logger.Any("owner_id", ownerID), logger.Any("run_id", runID),
			logger.String("reason", reason), logger.Error(err))
	}
}

func (s *Service) ListPayrollRuns(ctx context.Context, ownerID int64, params domain.ListPayrollRunsParams) ([]domain.PayrollRunRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.PayrollRunRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPayrollScope, ownerID, fmt.Sprintf("runs|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListPayrollRuns(ctx, ownerID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrPayrollTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) GetPayrollRun(ctx context.Context, ownerID, runID int64) (*domain.PayrollRunRecord, []domain.PayrollEntryRecord, error) {
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return nil, nil, err
	}

	entries, err := s.repo.ListPayrollEntries(ctx, ownerID, runID)
	if err != nil {
		return nil, nil, err
	}

	return run, entries, nil
}

// ListMyPayrollRuns lists payroll runs for the caller's own employeeID only,
// restricted to statuses that represent finalized pay. The employeeID must
// already be resolved from the authenticated user — never trust a
// client-supplied employee id for self-service scoping.
func (s *Service) ListMyPayrollRuns(ctx context.Context, ownerID, employeeID int64, params domain.ListPayrollRunsParams) ([]domain.PayrollRunRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListPayrollRunsForEmployee(ctx, ownerID, employeeID, params)
}

// GetMyPayrollRun returns a payroll run along with only the caller's own
// entry in it. If the employee has no entry in the run, or the run isn't in
// a self-service-visible status, it reports ErrPayrollRunNotFound rather than
// revealing that the run exists for other employees.
func (s *Service) GetMyPayrollRun(ctx context.Context, ownerID, runID, employeeID int64) (*domain.PayrollRunRecord, *domain.PayrollEntryRecord, error) {
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return nil, nil, err
	}
	if run.Status != domain.PayrollStatusApproved && run.Status != domain.PayrollStatusExported {
		return nil, nil, domain.ErrPayrollRunNotFound
	}

	entry, err := s.repo.GetPayrollEntryForEmployee(ctx, ownerID, runID, employeeID)
	if err != nil {
		return nil, nil, err
	}

	return run, entry, nil
}

func (s *Service) NotifyPayrollPeriodOpen(ctx context.Context) error {
	// Load every owner's payroll config in one query instead of listing owner
	// IDs and then fetching each config individually (N+1).
	configs, err := s.repo.ListAllPayrollConfigs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("payroll period notification: failed to list payroll configs", logger.Error(err))
		return err
	}

	today := int32(time.Now().Day()) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	for _, cfg := range configs {
		if today != cfg.PayDayOfMonth {
			continue
		}
		if dispErr := s.notifier.Dispatch(ctx, domain.NotificationEvent{
			OwnerID: cfg.OwnerID,
			Type:    "payroll.period_open",
			Subject: "Payroll period is open",
			Payload: map[string]any{
				"period_type": cfg.PeriodType,
				"pay_day":     cfg.PayDayOfMonth,
			},
		}); dispErr != nil {
			logger.Ctx(ctx).Warn("payroll period notification dispatch failed",
				logger.Any("owner_id", cfg.OwnerID),
				logger.Error(dispErr),
			)
		}
		logger.Ctx(ctx).Info("payroll period open notification sent",
			logger.Any("owner_id", cfg.OwnerID),
			logger.Any("period_type", cfg.PeriodType),
		)
	}

	return nil
}
