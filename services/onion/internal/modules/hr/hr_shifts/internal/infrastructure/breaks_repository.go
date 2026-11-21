package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func breakFromRow(row sqlc.HrBreak) domain.BreakRecord {
	rec := domain.BreakRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		StaffID:    row.StaffID,
		StaffName:  row.StaffName,
		BreakType:  row.BreakType,
		StartedAt:  row.StartedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.EndedAt.Valid {
		t := row.EndedAt.Time
		rec.EndedAt = &t
	}
	if row.DurationMinutes.Valid {
		d := row.DurationMinutes.Int32
		rec.DurationMinutes = &d
	}
	return rec
}

func (r *Repository) ListBreaks(ctx context.Context, businessID int64, params domain.ListBreaksParams) ([]domain.BreakRecord, error) {
	var empFilter sql.NullInt64
	if params.EmployeeID != nil && *params.EmployeeID > 0 {
		empFilter = sql.NullInt64{Int64: *params.EmployeeID, Valid: true}
	}
	rows, err := r.queries.ListBreaks(ctx, sqlc.ListBreaksParams{
		BusinessID: businessID,
		EmployeeID: empFilter,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list breaks", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.BreakRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, breakFromRow(row))
	}
	return out, nil
}

func (r *Repository) GetBreak(ctx context.Context, businessID, id int64) (*domain.BreakRecord, error) {
	row, err := r.queries.GetBreak(ctx, sqlc.GetBreakParams{BusinessID: businessID, ID: id})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBreakNotFound
		}
		logger.Ctx(ctx).Error("failed to get break", logger.Any("business_id", businessID), logger.Any("break_id", id), logger.Error(err))
		return nil, err
	}
	rec := breakFromRow(row)
	return &rec, nil
}

func (r *Repository) GetOpenBreakByEmployee(ctx context.Context, businessID, employeeID int64) (*domain.BreakRecord, error) {
	row, err := r.queries.GetOpenBreakByEmployee(ctx, sqlc.GetOpenBreakByEmployeeParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rec := breakFromRow(row)
	return &rec, nil
}

func (r *Repository) GetOpenBreakByStaff(ctx context.Context, businessID, staffID int64) (*domain.BreakRecord, error) {
	row, err := r.queries.GetOpenBreakByStaff(ctx, sqlc.GetOpenBreakByStaffParams{
		BusinessID: businessID,
		StaffID:    staffID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rec := breakFromRow(row)
	return &rec, nil
}

func (r *Repository) CreateBreak(ctx context.Context, businessID int64, input domain.StartBreakInput) (*domain.BreakRecord, error) {
	res, err := r.queries.CreateBreak(ctx, sqlc.CreateBreakParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: input.EmployeeID,
		StaffID:    input.StaffID,
		StaffName:  input.StaffName,
		BreakType:  input.BreakType,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create break", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetBreak(ctx, businessID, id)
}

func (r *Repository) EndBreak(ctx context.Context, businessID, id int64) (*domain.BreakRecord, error) {
	affected, err := r.queries.EndBreak(ctx, sqlc.EndBreakParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to end break", logger.Any("business_id", businessID), logger.Any("break_id", id), logger.Error(err))
		return nil, err
	}
	if affected == 0 {
		existing, getErr := r.GetBreak(ctx, businessID, id)
		if getErr != nil {
			return nil, getErr
		}
		if existing.EndedAt != nil {
			return nil, domain.ErrBreakAlreadyEnded
		}
		return nil, domain.ErrBreakNotFound
	}
	return r.GetBreak(ctx, businessID, id)
}
