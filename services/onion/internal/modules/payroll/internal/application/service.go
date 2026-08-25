package application

import (
	"context"
	"database/sql"
	"fmt"
	"math"
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

func invalidatePayrollScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(payrollScope, businessID))
	}
}

type repository interface {
	ListPayrollRecords(ctx context.Context, businessID int64, periodLabel string) ([]domain.PayrollRecord, error)
	GetCompletedShiftsForPeriod(ctx context.Context, businessID int64, start, end time.Time) ([]domain.ShiftSummary, error)
	GetStaffPayRates(ctx context.Context, businessID, staffID int64) (domain.StaffPayRates, error)
	GetStaffPayRatesByIDs(ctx context.Context, businessID int64, staffIDs []int64) (map[int64]domain.StaffPayRates, error)
	GetDefaultOvertimeMultiplier(ctx context.Context, businessID int64) (float64, bool, error)
	CreatePayrollRecords(ctx context.Context, businessID int64, records []domain.PayrollRecordInput) (int, error)
	MarkPayrollPaid(ctx context.Context, businessID, payrollID int64) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) ListPayrollRecords(ctx context.Context, businessID int64, periodLabel string) ([]domain.PayrollRecord, error) {
	key := cacheKeys.Scoped(payrollScope, businessID, fmt.Sprintf("list|p=%s", periodLabel))
	return cacher.Of[[]domain.PayrollRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.PayrollRecord, error) {
			return s.repo.ListPayrollRecords(ctx, businessID, periodLabel)
		},
		cacher.WithExpiry(ctx, payrollTTL),
	)
}

func (s *Service) RunPayroll(ctx context.Context, businessID int64, input domain.PayrollRunInput) (int, error) {
	periodStart, err := time.Parse("2006-01-02", input.PeriodStart)
	if err != nil {
		return 0, fmt.Errorf("invalid period_start date: %w", err)
	}

	periodEnd, err := time.Parse("2006-01-02", input.PeriodEnd)
	if err != nil {
		return 0, fmt.Errorf("invalid period_end date: %w", err)
	}

	summaries, err := s.repo.GetCompletedShiftsForPeriod(ctx, businessID, periodStart, periodEnd)
	if err != nil {
		return 0, err
	}

	records := make([]domain.PayrollRecordInput, 0, len(summaries))

	// Batch-load pay rates for all staff in the period in one query instead of
	// issuing GetStaffPayRates per summary (N+1).
	staffIDs := make([]int64, 0, len(summaries))
	for _, summary := range summaries {
		staffIDs = append(staffIDs, summary.StaffID)
	}
	ratesByStaff, err := s.repo.GetStaffPayRatesByIDs(ctx, businessID, staffIDs)
	if err != nil {
		return 0, err
	}

	// The overtime multiplier is tenant configuration owned by hr_overtime;
	// resolve it once per run instead of trusting a hardcoded default.
	overtimeMultiplier, configured, err := s.repo.GetDefaultOvertimeMultiplier(ctx, businessID)
	if err != nil {
		return 0, err
	}
	if !configured {
		overtimeMultiplier = 1.5
	}

	for _, summary := range summaries {
		rates, ok := ratesByStaff[summary.StaffID]
		if !ok {
			logger.Ctx(ctx).Error("Failed to get staff pay rates for payroll", logger.Any("staffId", summary.StaffID))
			continue
		}

		regularHours := summary.TotalHours - summary.TotalOvertimeHours
		if regularHours < 0 {
			regularHours = 0
		}

		var basePay, overtimePay float64

		if rates.HourlyRate > 0 {
			basePay = roundToCents(regularHours * rates.HourlyRate)
			overtimeRate := rates.OvertimeRate
			if overtimeRate == 0 {
				overtimeRate = rates.HourlyRate * overtimeMultiplier
			}
			overtimePay = roundToCents(summary.TotalOvertimeHours * overtimeRate)
		} else if rates.Salary > 0 {
			basePay = roundToCents(calculateBasePay(rates.Salary, periodStart, periodEnd))
		}

		netPay := roundToCents(basePay + overtimePay)

		records = append(records, domain.PayrollRecordInput{
			StaffID:       summary.StaffID,
			PeriodStart:   periodStart,
			PeriodEnd:     periodEnd,
			PeriodLabel:   input.PeriodLabel,
			RegularHours:  regularHours,
			OvertimeHours: summary.TotalOvertimeHours,
			BasePay:       basePay,
			OvertimePay:   overtimePay,
			NetPay:        netPay,
		})
	}

	createdCount, err := s.repo.CreatePayrollRecords(ctx, businessID, records)
	if err != nil {
		return 0, err
	}
	if createdCount > 0 {
		invalidatePayrollScope(ctx, businessID)
	}
	return createdCount, nil
}

func (s *Service) MarkPayrollPaid(ctx context.Context, businessID, payrollID int64) error {
	if err := s.repo.MarkPayrollPaid(ctx, businessID, payrollID); err != nil {
		return err
	}
	invalidatePayrollScope(ctx, businessID)
	return nil
}

// roundToCents rounds a money amount to whole cents once, at the last point
// before persistence, so accumulated float error cannot leak into DECIMAL
// columns.
func roundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// calculateBasePay prorates an annual salary over an inclusive calendar date
// range. Each month segment contributes salary x segmentDays / daysInYear,
// so a complete January pays 31 days (not the 30 an exclusive subtraction
// yields) and a complete leap year sums to exactly one salary.
func calculateBasePay(annualSalary float64, startDate, endDate time.Time) float64 {
	if annualSalary <= 0 || endDate.Before(startDate) {
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
		basePay += annualSalary * inclusiveDays / float64(daysInYear(current.Year()))
		current = segmentEnd.AddDate(0, 0, 1)
	}

	return basePay
}

func daysInYear(year int) int {
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return 366
	}
	return 365
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}
