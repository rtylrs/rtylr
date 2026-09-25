package application

import (
	"context"
	"database/sql"
	"errors"
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
		Status:          domain.PayrollStatusInReview,
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
