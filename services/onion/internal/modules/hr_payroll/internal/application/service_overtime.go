package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

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
	if run.Status != domain.PayrollStatusInReview {
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
