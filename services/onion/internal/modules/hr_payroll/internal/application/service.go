package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
}

func NewService(ctx context.Context, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(ctx),
		approvalProvider: approvalProvider,
	}
}

func (s *Service) GetPayrollConfig(ctx context.Context, ownerID int64) (*domain.PayrollConfigRecord, error) {
	cfg, err := s.repo.GetPayrollConfig(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) UpsertPayrollConfig(ctx context.Context, ownerID int64, cfg domain.PayrollConfigRecord) error {
	if cfg.PeriodType == "" {
		cfg.PeriodType = "monthly"
	}
	if cfg.PayDayOfMonth <= 0 || cfg.PayDayOfMonth > 28 {
		cfg.PayDayOfMonth = 28
	}
	return s.repo.UpsertPayrollConfig(ctx, ownerID, cfg)
}

func (s *Service) TriggerPayrollRun(ctx context.Context, ownerID, createdBy int64, periodStart, periodEnd string) (*domain.PayrollRunRecord, []domain.PayrollEntryRecord, error) {
	startDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, nil, domain.ErrInvalidDateRange
	}
	if !endDate.After(startDate) {
		return nil, nil, domain.ErrInvalidDateRange
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

	periodDays := endDate.Sub(startDate).Hours() / 24

	var entries []domain.PayrollEntryRecord
	var totalGross, totalNet float64

	for _, emp := range employees {
		hoursWorked := float64(tsMap[emp.EmployeeID]) / 60.0
		overtimeMinutes := int32(0)
		multiplier := defaultMultiplier

		if ot, ok := otMap[emp.EmployeeID]; ok {
			overtimeMinutes = ot.TotalMinutes
			if ot.Multiplier > 0 {
				multiplier = ot.Multiplier
			}
		}
		overtimeHours := float64(overtimeMinutes) / 60.0

		var basePay float64
		switch emp.PayType {
		case "hourly":
			basePay = hoursWorked * emp.HourlyRate
		case "monthly":
			daysInMonth := 30.0
			basePay = emp.MonthlySalary * (periodDays / daysInMonth)
		default:
			basePay = emp.MonthlySalary * (periodDays / 30.0)
		}

		overtimePay := 0.0
		if cfg.OvertimeEnabled && overtimeHours > 0 && emp.HourlyRate > 0 {
			overtimePay = overtimeHours * emp.HourlyRate * multiplier
		}

		netPay := basePay + overtimePay
		grossPay := netPay

		entry := domain.PayrollEntryRecord{
			OwnerID:       ownerID,
			EmployeeID:    emp.EmployeeID,
			HoursWorked:   hoursWorked,
			OvertimeHours: overtimeHours,
			BasePay:       basePay,
			OvertimePay:   overtimePay,
			Bonuses:       0,
			Deductions:    0,
			NetPay:        netPay,
		}

		entries = append(entries, entry)
		totalGross += grossPay
		totalNet += netPay
	}

	run := domain.PayrollRunRecord{
		OwnerID:     ownerID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      "draft",
		CreatedBy:   createdBy,
		TotalGross:  totalGross,
		TotalNet:    totalNet,
	}

	createdRun, createdEntries, err := s.repo.CreatePayrollRunWithEntries(ctx, ownerID, run, entries)
	if err != nil {
		return nil, nil, err
	}

	return createdRun, createdEntries, nil
}

func (s *Service) ApprovePayrollRun(ctx context.Context, ownerID, runID, approverID int64) error {
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return err
	}

	chain, err := s.approvalProvider.GetApprovalChain(ctx, ownerID, "payroll")
	if err != nil {
		return err
	}
	if chain == nil {
		return domain.ErrApprovalChainNotConfigured
	}

	if run.Status == "draft" {
		if err := s.repo.UpdatePayrollRunStatus(ctx, ownerID, runID, "in_review"); err != nil {
			return err
		}
	}

	return s.repo.SetPayrollRunApproved(ctx, ownerID, runID, approverID)
}

func (s *Service) ExportPayrollToFinance(ctx context.Context, ownerID, runID int64) error {
	run, err := s.repo.GetPayrollRun(ctx, ownerID, runID)
	if err != nil {
		return err
	}

	if run.Status != "approved" {
		return domain.ErrPayrollRunNotApproved
	}

	if !run.ExportedAt.IsZero() {
		return domain.ErrPayrollAlreadyExported
	}

	// TODO: Call Finance module's internal API to create journal entries
	// Interface: financeService.CreatePayrollJournalEntry(ctx, ownerID, runID, run.TotalGross, run.TotalNet)
	// This would create debit/credit entries in the finance_journal_entries table

	if err := s.repo.SetPayrollRunExported(ctx, ownerID, runID); err != nil {
		return err
	}

	// TODO: Dispatch pay slip notifications per employee via notification service
	// Interface: notificationService.Dispatch(ctx, ownerID, domain.NotificationEvent{Type: "payslip", RunID: runID})

	return nil
}

func (s *Service) ListPayrollRuns(ctx context.Context, ownerID int64, params domain.ListPayrollRunsParams) ([]domain.PayrollRunRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListPayrollRuns(ctx, ownerID, params)
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

func (s *Service) NotifyPayrollPeriodOpen(ctx context.Context) error {
	ownerIDs, err := s.repo.ListDistinctOwnerIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("payroll period notification: failed to list owner IDs", logger.Error(err))
		return err
	}

	for _, ownerID := range ownerIDs {
		cfg, cfgErr := s.repo.GetPayrollConfig(ctx, ownerID)
		if cfgErr != nil || cfg == nil {
			continue
		}

		now := time.Now()
		if int32(now.Day()) == cfg.PayDayOfMonth {
			// TODO: Dispatch notification to Finance Manager via notification service
			// Interface: notificationService.Dispatch(ctx, ownerID, domain.NotificationEvent{Type: "payroll_period_open"})
			logger.Ctx(ctx).Info("payroll period open notification sent",
				logger.Any("owner_id", ownerID),
				logger.Any("period_type", cfg.PeriodType),
			)
		}
	}

	return nil
}
