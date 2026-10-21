package infrastructure

import (
	"context"
	"database/sql"
	"time"

	hrcontract "github.com/voxire/onion/internal/contracts/hr"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// The reads below serve Insights reports, so Insights never queries HR
// tables itself. Dates are compared as calendar days.

func (r *Repository) InsightsActiveEmployees(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountActiveEmployees(ctx, businessID)
}

func (r *Repository) InsightsActiveStaff(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountActiveStaff(ctx, businessID)
}

func (r *Repository) InsightsShiftsThisWeek(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountShiftsThisWeek(ctx, businessID)
}

func (r *Repository) InsightsPendingLeave(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountPendingLeave(ctx, businessID)
}

// InsightsOpenLeave counts leave requests that are not rejected, cancelled or
// completed.
func (r *Repository) InsightsOpenLeave(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountOpenLeave(ctx, businessID)
}

// InsightsHeadcountAsOf counts employees whose employment covers asOf.
func (r *Repository) InsightsHeadcountAsOf(ctx context.Context, businessID int64, asOf time.Time) (int64, error) {
	return r.queries.InsightsHeadcountAsOf(ctx, sqlc.InsightsHeadcountAsOfParams{BusinessID: businessID, AsOf: sql.NullTime{Time: asOf, Valid: true}})
}

func (r *Repository) InsightsHires(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountHires(ctx, sqlc.InsightsCountHiresParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: true},
	})
}

func (r *Repository) InsightsDepartures(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountDepartures(ctx, sqlc.InsightsCountDeparturesParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     sql.NullTime{Time: to, Valid: true},
	})
}

func (r *Repository) InsightsShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (hrcontract.ShiftMetrics, error) {
	row, err := r.queries.InsightsShiftMetrics(ctx, sqlc.InsightsShiftMetricsParams{BusinessID: businessID, DateFrom: from, DateTo: to})
	if err != nil {
		return hrcontract.ShiftMetrics{}, err
	}
	return hrcontract.ShiftMetrics{
		Scheduled: int64(sqlutil.DecimalToFloat(row.Scheduled)),
		Completed: int64(sqlutil.DecimalToFloat(row.Completed)),
		NoShows:   int64(sqlutil.DecimalToFloat(row.NoShows)),
	}, nil
}

// InsightsLeaveStarting counts leave requests in a status whose start date
// falls in the range.
func (r *Repository) InsightsLeaveStarting(ctx context.Context, businessID int64, status string, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountLeaveStarting(ctx, sqlc.InsightsCountLeaveStartingParams{BusinessID: businessID, Status: status, DateFrom: from, DateTo: to})
}

// InsightsScheduledHours sums the hours of non-cancelled shift assignments.
func (r *Repository) InsightsScheduledHours(ctx context.Context, businessID int64, from, to time.Time) (float64, error) {
	hours, err := r.queries.InsightsScheduledHours(ctx, sqlc.InsightsScheduledHoursParams{BusinessID: businessID, DateFrom: from, DateTo: to})
	if err != nil {
		return 0, err
	}
	return sqlutil.DecimalToFloat(hours), nil
}

func (r *Repository) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.InsightsHRLastChange(ctx, sqlc.InsightsHRLastChangeParams{BusinessID: businessID})
}
