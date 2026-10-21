package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/labourcost"
	"github.com/voxire/onion/internal/platform/logger"
)

// cross-schema read from hr_overtime — approved overtime minutes with multipliers for payroll computation.
func (r *Repository) ListApprovedOvertimeSummaries(ctx context.Context, ownerID int64, periodStart, periodEnd string) ([]domain.OvertimeSummary, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListApprovedOvertimeSummaries(ctx, sqlc.ListApprovedOvertimeSummariesParams{
		BusinessID: ownerID,
		Date:       ps,
		Date_2:     pe,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime summaries", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	summaries := make([]domain.OvertimeSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.OvertimeSummary{
			EmployeeID:        row.EmployeeID,
			TotalMinutes:      interfaceToInt32(row.TotalMinutes),
			WeightedMinutes:   interfaceToFloat64(row.WeightedMinutes),
			UnweightedMinutes: interfaceToFloat64(row.UnweightedMinutes),
		})
	}
	return summaries, nil
}

// cross-schema read from hr_core — employee contract type, rates, and schedule for payroll computation.
func (r *Repository) ListEmployeePayInfo(ctx context.Context, ownerID int64) ([]domain.EmployeePayInfo, error) {
	rows, err := r.queries.ListEmployeePayInfo(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employee pay info", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	employees := make([]domain.EmployeePayInfo, 0, len(rows))
	for _, row := range rows {
		hourlyRateCents := nullDecimalStringToCents(row.HourlyRate)
		salaryCents := nullDecimalStringToCents(row.Salary)

		payType := "monthly"
		if hourlyRateCents > 0 && salaryCents == 0 {
			payType = "hourly"
		}

		employees = append(employees, domain.EmployeePayInfo{
			EmployeeID:         row.ID,
			PayType:            payType,
			HourlyRateCents:    hourlyRateCents,
			MonthlySalaryCents: salaryCents,
		})
	}
	return employees, nil
}

// GetProjectLabourCost is the only HR-owned read exposed to project-cost
// consumers. It aggregates billable minutes and rates inside HR and returns no
// employee compensation fields.
func (r *Repository) GetProjectLabourCost(ctx context.Context, businessID, projectID int64) (labourcost.ProjectCost, error) {
	row, err := r.queries.GetProjectLabourCost(ctx, sqlc.GetProjectLabourCostParams{
		BusinessID:   businessID,
		BusinessID_2: businessID,
		ProjectID:    sql.NullInt64{Int64: projectID, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		// The aggregate groups by project, so a project with no billable time
		// yields no row. That is a zero cost, not a failure.
		return labourcost.ProjectCost{BusinessID: businessID, ProjectID: projectID}, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("failed to get project labour cost",
			logger.Any("business_id", businessID), logger.Any("project_id", projectID), logger.Error(err))
		return labourcost.ProjectCost{}, err
	}
	return labourcost.ProjectCost{
		BusinessID:            row.BusinessID,
		ProjectID:             row.ProjectID.Int64,
		CostCents:             interfaceToInt64(row.CostCents),
		CostedMinutes:         interfaceToInt64(row.CostedMinutes),
		UncostedMinutes:       interfaceToInt64(row.UncostedMinutes),
		UncostedEmployeeCount: row.UncostedEmployeeCount,
	}, nil
}

// cross-schema read from hr_overtime — overtime rate multiplier configuration for payroll computation.
func (r *Repository) GetOvertimeConfig(ctx context.Context, ownerID int64) (float64, bool, error) {
	row, err := r.queries.GetOvertimeConfigForPayroll(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return parseFloat(row.DefaultMultiplier), true, nil
}

func (r *Repository) ListDistinctOwnerIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListDistinctPayrollOwnerIDs(ctx)
}

// ListAllPayrollConfigs returns every payroll config in one query. It exists so
// the period-open notification job can load all owners' configs at once instead
// of listing owner IDs and then fetching each config individually (N+1).
func (r *Repository) ListAllPayrollConfigs(ctx context.Context) ([]domain.PayrollConfigRecord, error) {
	rows, err := r.queries.ListAllPayrollConfigs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list all payroll configs", logger.Error(err))
		return nil, err
	}

	var configs []domain.PayrollConfigRecord
	for _, row := range rows {
		configs = append(configs, payrollConfigFromRow(row))
	}

	return configs, nil
}

// --- Row Mappers ---

func payrollConfigFromRow(row sqlc.HrPayrollConfig) domain.PayrollConfigRecord {
	return domain.PayrollConfigRecord{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		PeriodType:      row.PeriodType,
		PayDayOfMonth:   row.PayDayOfMonth,
		OvertimeEnabled: row.OvertimeEnabled,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func payrollRunFromRow(row sqlc.HrPayrollRun) domain.PayrollRunRecord {
	run := domain.PayrollRunRecord{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		PeriodStart:     row.PeriodStart.Format("2006-01-02"),
		PeriodEnd:       row.PeriodEnd.Format("2006-01-02"),
		Status:          row.Status,
		CreatedBy:       row.CreatedBy,
		ApprovedBy:      row.ApprovedBy.Int64,
		TotalGrossCents: decimalStringToCentsOrZero(row.TotalGross),
		TotalNetCents:   decimalStringToCentsOrZero(row.TotalNet),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.ExportedAt.Valid {
		run.ExportedAt = row.ExportedAt.Time
	}
	return run
}

func payrollEntryFromRow(row sqlc.HrPayrollEntry) domain.PayrollEntryRecord {
	return domain.PayrollEntryRecord{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		PayrollRunID:     row.PayrollRunID,
		EmployeeID:       row.EmployeeID,
		HoursWorked:      parseFloat(row.HoursWorked),
		OvertimeHours:    parseFloat(row.OvertimeHours),
		BasePayCents:     decimalStringToCentsOrZero(row.BasePay),
		OvertimePayCents: decimalStringToCentsOrZero(row.OvertimePay),
		BonusesCents:     decimalStringToCentsOrZero(row.Bonuses),
		DeductionsCents:  decimalStringToCentsOrZero(row.Deductions),
		NetPayCents:      decimalStringToCentsOrZero(row.NetPay),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

// --- Helpers ---

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return f
}

// centsToDecimalString renders an integer cents amount as the DECIMAL(15,2)
// string sqlc expects, using only integer arithmetic so no float64 rounding
// error can be introduced on the way into the database.
func centsToDecimalString(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole := cents / 100
	fraction := cents % 100
	s := strconv.FormatInt(whole, 10) + "." + fmt.Sprintf("%02d", fraction)
	if negative {
		s = "-" + s
	}
	return s
}
