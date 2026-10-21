package hr_reports

import (
	"context"
	"time"

	hrcontract "github.com/voxire/onion/internal/contracts/hr"
)

// Insights reports read HR data through these methods rather than querying HR
// tables.

func (m *Module) InsightsActiveEmployees(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsActiveEmployees(ctx, businessID)
}

func (m *Module) InsightsActiveStaff(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsActiveStaff(ctx, businessID)
}

func (m *Module) InsightsShiftsThisWeek(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsShiftsThisWeek(ctx, businessID)
}

func (m *Module) InsightsPendingLeave(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsPendingLeave(ctx, businessID)
}

func (m *Module) InsightsOpenLeave(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsOpenLeave(ctx, businessID)
}

func (m *Module) InsightsHeadcountAsOf(ctx context.Context, businessID int64, asOf time.Time) (int64, error) {
	return m.repository.InsightsHeadcountAsOf(ctx, businessID, asOf)
}

func (m *Module) InsightsHires(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsHires(ctx, businessID, from, to)
}

func (m *Module) InsightsDepartures(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsDepartures(ctx, businessID, from, to)
}

func (m *Module) InsightsShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (hrcontract.ShiftMetrics, error) {
	return m.repository.InsightsShiftMetrics(ctx, businessID, from, to)
}

func (m *Module) InsightsLeaveStarting(ctx context.Context, businessID int64, status string, from, to time.Time) (int64, error) {
	return m.repository.InsightsLeaveStarting(ctx, businessID, status, from, to)
}

func (m *Module) InsightsScheduledHours(ctx context.Context, businessID int64, from, to time.Time) (float64, error) {
	return m.repository.InsightsScheduledHours(ctx, businessID, from, to)
}

func (m *Module) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return m.repository.InsightsLastChange(ctx, businessID)
}
