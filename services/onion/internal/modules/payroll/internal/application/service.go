package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	payrollScope = "payroll"
	payrollTTL   = 5 * time.Minute
)

func invalidatePayrollScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(payrollScope, ownerID))
	}
}

type repository interface {
	ListPayrollRecords(ctx context.Context, ownerID int64, periodLabel string) ([]domain.PayrollRecord, error)
	GetCompletedShiftsForPeriod(ctx context.Context, ownerID int64, start, end time.Time) ([]domain.ShiftSummary, error)
	GetStaffPayRates(ctx context.Context, ownerID, staffID int64) (domain.StaffPayRates, error)
	CreatePayrollRecord(ctx context.Context, ownerID, staffID int64, periodStart, periodEnd time.Time, periodLabel string, regularHours, overtimeHours, basePay, overtimePay, netPay float64) (int64, error)
	MarkPayrollPaid(ctx context.Context, ownerID, payrollID int64) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListPayrollRecords(ctx context.Context, ownerID int64, periodLabel string) ([]domain.PayrollRecord, error) {
	key := cacheKeys.Scoped(payrollScope, ownerID, fmt.Sprintf("list|p=%s", periodLabel))
	return cacher.Of[[]domain.PayrollRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.PayrollRecord, error) {
			return s.repo.ListPayrollRecords(ctx, ownerID, periodLabel)
		},
		cacher.WithExpiry(ctx, payrollTTL),
	)
}

func (s *Service) RunPayroll(ctx context.Context, ownerID int64, input domain.PayrollRunInput) (int, error) {
	periodStart, err := time.Parse("2006-01-02", input.PeriodStart)
	if err != nil {
		return 0, fmt.Errorf("invalid period_start date: %w", err)
	}

	periodEnd, err := time.Parse("2006-01-02", input.PeriodEnd)
	if err != nil {
		return 0, fmt.Errorf("invalid period_end date: %w", err)
	}

	summaries, err := s.repo.GetCompletedShiftsForPeriod(ctx, ownerID, periodStart, periodEnd)
	if err != nil {
		return 0, err
	}

	createdCount := 0

	for _, summary := range summaries {
		rates, rateErr := s.repo.GetStaffPayRates(ctx, ownerID, summary.StaffID)
		if rateErr != nil {
			logger.Ctx(ctx).Error("Failed to get staff pay rates for payroll", logger.Error(rateErr), logger.Any("staffId", summary.StaffID))
			continue
		}

		regularHours := summary.TotalHours - summary.TotalOvertimeHours
		if regularHours < 0 {
			regularHours = 0
		}

		var basePay, overtimePay float64

		if rates.HourlyRate > 0 {
			basePay = regularHours * rates.HourlyRate
			overtimeRate := rates.OvertimeRate
			if overtimeRate == 0 {
				overtimeRate = rates.HourlyRate * 1.5
			}
			overtimePay = summary.TotalOvertimeHours * overtimeRate
		} else if rates.Salary > 0 {
			totalDays := periodEnd.Sub(periodStart).Hours() / 24
			if totalDays > 0 {
				basePay = (rates.Salary / 365) * totalDays
			}
		}

		netPay := basePay + overtimePay

		_, createErr := s.repo.CreatePayrollRecord(ctx, ownerID, summary.StaffID, periodStart, periodEnd,
			input.PeriodLabel, regularHours, summary.TotalOvertimeHours, basePay, overtimePay, netPay)
		if createErr != nil {
			logger.Ctx(ctx).Error("Failed to create payroll record", logger.Error(createErr), logger.Any("staffId", summary.StaffID))
			continue
		}

		createdCount++
	}

	if createdCount > 0 {
		invalidatePayrollScope(ctx, ownerID)
	}
	return createdCount, nil
}

func (s *Service) MarkPayrollPaid(ctx context.Context, ownerID, payrollID int64) error {
	if err := s.repo.MarkPayrollPaid(ctx, ownerID, payrollID); err != nil {
		return err
	}
	invalidatePayrollScope(ctx, ownerID)
	return nil
}
