package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// ScheduledReportRow is the repository-level shape of one row in
// insights_scheduled_reports. All time fields are normalised to UTC
// before being returned from the service layer.
type ScheduledReportRow struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ErrScheduledReportNotFound mirrors the sentinel-style error pattern used
// across every other module.
var ErrScheduledReportNotFound = errors.New("scheduled report not found")

// ListScheduledReports returns every scheduled report the owner has
// configured. The list is not paginated at P3 because a realistic upper
// bound is low (tens of schedules per tenant).
func (r *Repository) ListScheduledReports(ctx context.Context, businessID int64) ([]ScheduledReportRow, error) {
	const q = `
		SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
		FROM insights_scheduled_reports
		WHERE business_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: list scheduled reports failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make([]ScheduledReportRow, 0)
	for rows.Next() {
		var sr ScheduledReportRow
		if scanErr := rows.Scan(
			&sr.ID, &sr.OwnerID, &sr.BusinessID, &sr.TemplateID, &sr.Cadence,
			&sr.Channel, &sr.Recipient, &sr.IsActive,
			&sr.CreatedAt, &sr.UpdatedAt,
		); scanErr != nil {
			logger.Ctx(ctx).Error("insights: list scheduled reports scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		out = append(out, sr)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}

// CreateScheduledReport inserts a new schedule and returns the assigned id.
func (r *Repository) CreateScheduledReport(ctx context.Context, row ScheduledReportRow) (int64, error) {
	const q = `
		INSERT INTO insights_scheduled_reports
		  (owner_id, business_id, template_id, cadence, channel, recipient, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.pool.ExecContext(ctx, q,
		row.OwnerID, row.BusinessID, row.TemplateID, row.Cadence, row.Channel, row.Recipient, row.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("insights: create scheduled report failed", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateScheduledReport updates the mutable fields. business_id is part of
// the WHERE clause to prevent cross-tenant writes.
func (r *Repository) UpdateScheduledReport(ctx context.Context, row ScheduledReportRow) error {
	const q = `
		UPDATE insights_scheduled_reports
		SET template_id = ?, cadence = ?, channel = ?, recipient = ?, is_active = ?
		WHERE id = ? AND business_id = ?
	`
	res, err := r.pool.ExecContext(ctx, q,
		row.TemplateID, row.Cadence, row.Channel, row.Recipient, row.IsActive,
		row.ID, row.BusinessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("insights: update scheduled report failed", logger.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrScheduledReportNotFound
	}
	return nil
}

// DeleteScheduledReport removes a schedule; safe for missing rows (returns
// ErrScheduledReportNotFound rather than silently succeeding so the
// caller can translate that into a 404).
func (r *Repository) DeleteScheduledReport(ctx context.Context, businessID, id int64) error {
	res, err := r.pool.ExecContext(ctx,
		`DELETE FROM insights_scheduled_reports WHERE id = ? AND business_id = ?`,
		id, businessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("insights: delete scheduled report failed", logger.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrScheduledReportNotFound
	}
	return nil
}

// GetScheduledReport reads one row; used by the PUT/DELETE paths to
// surface a 404 before doing the mutation.
func (r *Repository) GetScheduledReport(ctx context.Context, businessID, id int64) (ScheduledReportRow, error) {
	const q = `
		SELECT id, owner_id, business_id, template_id, cadence, channel, recipient, is_active, created_at, updated_at
		FROM insights_scheduled_reports
		WHERE id = ? AND business_id = ?
	`
	row := r.pool.QueryRowContext(ctx, q, id, businessID)
	var sr ScheduledReportRow
	if err := row.Scan(&sr.ID, &sr.OwnerID, &sr.BusinessID, &sr.TemplateID, &sr.Cadence, &sr.Channel, &sr.Recipient, &sr.IsActive, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScheduledReportRow{}, ErrScheduledReportNotFound
		}
		logger.Ctx(ctx).Error("insights: get scheduled report failed", logger.Error(err))
		return ScheduledReportRow{}, err
	}
	return sr, nil
}
