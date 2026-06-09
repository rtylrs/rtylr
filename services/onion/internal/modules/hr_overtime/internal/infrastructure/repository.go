package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListOvertimeRequests(ctx context.Context, ownerID int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error) {
	total, err := r.queries.CountOvertimeRequests(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count overtime requests", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListOvertimeRequests(ctx, sqlc.ListOvertimeRequestsParams{
		OwnerID: ownerID,
		Limit:   params.Limit,
		Offset:  params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime requests", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	requests := make([]domain.OvertimeRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, overtimeRequestFromRow(row))
	}
	return requests, total, nil
}

func (r *Repository) GetOvertimeRequest(ctx context.Context, ownerID, requestID int64) (*domain.OvertimeRequestRecord, error) {
	row, err := r.queries.GetOvertimeRequest(ctx, sqlc.GetOvertimeRequestParams{OwnerID: ownerID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOvertimeRequestNotFound
		}
		logger.Ctx(ctx).Error("failed to get overtime request", logger.Any("owner_id", ownerID), logger.Any("request_id", requestID), logger.Error(err))
		return nil, err
	}
	req := overtimeRequestFromRow(row)
	return &req, nil
}

func (r *Repository) CreateOvertimeRequest(ctx context.Context, ownerID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	reqDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateOvertimeRequest(ctx, sqlc.CreateOvertimeRequestParams{
		OwnerID:              ownerID,
		EmployeeID:           req.EmployeeID,
		Date:                 reqDate,
		PlannedExtraMinutes:  req.PlannedExtraMinutes,
		Reason:               nullString(req.Reason),
		Status:               req.Status,
		CurrentApprovalLevel: req.CurrentApprovalLevel,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create overtime request", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	req.ID = id
	req.OwnerID = ownerID
	return &req, nil
}

func (r *Repository) ApproveOvertimeRequest(ctx context.Context, ownerID, requestID, approverID int64, multiplier float64) error {
	return r.queries.ApproveOvertimeRequest(ctx, sqlc.ApproveOvertimeRequestParams{
		ApprovedBy:        sql.NullInt64{Int64: approverID, Valid: approverID != 0},
		MultiplierApplied: nullStringFromFloat(multiplier),
		ID:                requestID,
		OwnerID:           ownerID,
	})
}

func (r *Repository) AdvanceOvertimeApprovalLevel(ctx context.Context, ownerID, requestID int64) error {
	return r.queries.AdvanceOvertimeApprovalLevel(ctx, sqlc.AdvanceOvertimeApprovalLevelParams{ID: requestID, OwnerID: ownerID})
}

func (r *Repository) RejectOvertimeRequest(ctx context.Context, ownerID, requestID int64) error {
	return r.queries.RejectOvertimeRequest(ctx, sqlc.RejectOvertimeRequestParams{ID: requestID, OwnerID: ownerID})
}

func (r *Repository) CreateFlaggedOvertimeRequest(ctx context.Context, ownerID, employeeID int64, date string, extraMinutes int32) error {
	reqDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}
	_, err = r.queries.CreateFlaggedOvertimeRequest(ctx, sqlc.CreateFlaggedOvertimeRequestParams{
		OwnerID:             ownerID,
		EmployeeID:          employeeID,
		Date:                reqDate,
		PlannedExtraMinutes: extraMinutes,
	})
	return err
}

func (r *Repository) ListFlaggedForDate(ctx context.Context, ownerID int64, date string) ([]domain.OvertimeRequestRecord, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListFlaggedForDate(ctx, sqlc.ListFlaggedForDateParams{OwnerID: ownerID, Date: d})
	if err != nil {
		return nil, err
	}
	requests := make([]domain.OvertimeRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, overtimeRequestFromRow(row))
	}
	return requests, nil
}

func (r *Repository) GetOvertimeConfig(ctx context.Context, ownerID int64) (*domain.OvertimeConfigRecord, error) {
	row, err := r.queries.GetOvertimeConfig(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOvertimeConfigNotFound
		}
		return nil, err
	}
	cfg := overtimeConfigFromRow(row)
	return &cfg, nil
}

func (r *Repository) UpsertOvertimeConfig(ctx context.Context, ownerID int64, cfg domain.OvertimeConfigRecord) error {
	_, err := r.queries.UpsertOvertimeConfig(ctx, sqlc.UpsertOvertimeConfigParams{
		OwnerID:               ownerID,
		DailyThresholdMinutes: cfg.DailyThresholdMinutes,
		DefaultMultiplier:     fmt.Sprintf("%.2f", cfg.DefaultMultiplier),
		AutoDetectEnabled:     cfg.AutoDetectEnabled,
	})
	return err
}

func (r *Repository) ListAllConfiguredOwnerIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListAllConfiguredOwnerIDs(ctx)
}

// --- Helpers ---

func overtimeRequestFromRow(row sqlc.HrOvertimeRequest) domain.OvertimeRequestRecord {
	return domain.OvertimeRequestRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		EmployeeID:           row.EmployeeID,
		Date:                 row.Date.Format("2006-01-02"),
		PlannedExtraMinutes:  row.PlannedExtraMinutes,
		Reason:               row.Reason.String,
		Status:               row.Status,
		CurrentApprovalLevel: row.CurrentApprovalLevel,
		ApprovedBy:           row.ApprovedBy.Int64,
		MultiplierApplied:    parseFloat(row.MultiplierApplied.String),
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func overtimeConfigFromRow(row sqlc.HrOvertimeConfig) domain.OvertimeConfigRecord {
	return domain.OvertimeConfigRecord{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		DailyThresholdMinutes: row.DailyThresholdMinutes,
		DefaultMultiplier:     parseFloat(row.DefaultMultiplier),
		AutoDetectEnabled:     row.AutoDetectEnabled,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse overtime decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

func nullStringFromFloat(f float64) sql.NullString {
	if f == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", f), Valid: true}
}
