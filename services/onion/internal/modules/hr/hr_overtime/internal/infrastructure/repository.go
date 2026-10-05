package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_overtime/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) ListOvertimeRequests(ctx context.Context, businessID int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error) {
	employeeID := sql.NullInt64{}
	if params.EmployeeID != nil {
		employeeID = sql.NullInt64{Int64: *params.EmployeeID, Valid: true}
	}
	status := sql.NullString{}
	if params.Status != nil {
		status = sql.NullString{String: *params.Status, Valid: true}
	}

	filterParams := sqlc.CountOvertimeRequestsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     status,
	}
	total, err := r.queries.CountOvertimeRequests(ctx, filterParams)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count overtime requests", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListOvertimeRequests(ctx, sqlc.ListOvertimeRequestsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Status:     status,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime requests", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	requests := make([]domain.OvertimeRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, overtimeRequestFromRow(row))
	}
	return requests, total, nil
}

func (r *Repository) GetOvertimeRequest(ctx context.Context, businessID, requestID int64) (*domain.OvertimeRequestRecord, error) {
	row, err := r.queries.GetOvertimeRequest(ctx, sqlc.GetOvertimeRequestParams{BusinessID: businessID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOvertimeRequestNotFound
		}
		logger.Ctx(ctx).Error("failed to get overtime request", logger.Any("business_id", businessID), logger.Any("request_id", requestID), logger.Error(err))
		return nil, err
	}
	req := overtimeRequestFromRow(row)
	return &req, nil
}

func (r *Repository) CreateOvertimeRequest(ctx context.Context, businessID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	reqDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateOvertimeRequest(ctx, sqlc.CreateOvertimeRequestParams{
		OwnerID:              businessID,
		BusinessID:           businessID,
		EmployeeID:           req.EmployeeID,
		Date:                 reqDate,
		PlannedExtraMinutes:  req.PlannedExtraMinutes,
		Reason:               sqlutil.NullString(req.Reason),
		Status:               req.Status,
		CurrentApprovalLevel: req.CurrentApprovalLevel,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create overtime request", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	req.ID = id
	req.OwnerID = businessID
	req.BusinessID = businessID

	// Re-read the row after INSERT. MySQL assigns created_at/updated_at in the
	// database, so returning the pre-insert struct serialized Go's zero time
	// as a large negative Unix timestamp. The scoped read also makes the
	// returned record authoritative for any future database defaults.
	return r.GetOvertimeRequest(ctx, businessID, id)
}

func (r *Repository) ApproveOvertimeRequest(ctx context.Context, businessID, requestID, approverID int64, multiplier float64) error {
	result, err := r.queries.ApproveOvertimeRequest(ctx, sqlc.ApproveOvertimeRequestParams{
		ApprovedBy:        sql.NullInt64{Int64: approverID, Valid: approverID != 0},
		MultiplierApplied: nullStringFromFloat(multiplier),
		ID:                requestID,
		BusinessID:        businessID,
	})
	if err != nil {
		return err
	}
	return overtimeRequestWriteResult(result)
}

func (r *Repository) AdvanceOvertimeApprovalLevel(ctx context.Context, businessID, requestID int64, currentLevel int32) error {
	result, err := r.queries.AdvanceOvertimeApprovalLevel(ctx, sqlc.AdvanceOvertimeApprovalLevelParams{
		ID: requestID, BusinessID: businessID, CurrentApprovalLevel: currentLevel,
	})
	if err != nil {
		return err
	}
	return overtimeRequestWriteResult(result)
}

func (r *Repository) RejectOvertimeRequest(ctx context.Context, businessID, requestID int64, rejectionReason string) error {
	result, err := r.queries.RejectOvertimeRequest(ctx, sqlc.RejectOvertimeRequestParams{
		ID:              requestID,
		BusinessID:      businessID,
		RejectionReason: sqlutil.NullString(rejectionReason),
	})
	if err != nil {
		return err
	}
	return overtimeRequestWriteResult(result)
}

func (r *Repository) CreateFlaggedOvertimeRequest(ctx context.Context, businessID, employeeID int64, date string, extraMinutes int32) error {
	reqDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}
	_, err = r.queries.CreateFlaggedOvertimeRequest(ctx, sqlc.CreateFlaggedOvertimeRequestParams{
		OwnerID:             businessID,
		BusinessID:          businessID,
		EmployeeID:          employeeID,
		Date:                reqDate,
		PlannedExtraMinutes: extraMinutes,
		BusinessID_2:        businessID,
		EmployeeID_2:        employeeID,
		Date_2:              reqDate,
	})
	return err
}

func (r *Repository) ListFlaggedForDate(ctx context.Context, businessID int64, date string) ([]domain.OvertimeRequestRecord, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListFlaggedForDate(ctx, sqlc.ListFlaggedForDateParams{BusinessID: businessID, Date: d})
	if err != nil {
		return nil, err
	}
	requests := make([]domain.OvertimeRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, overtimeRequestFromRow(row))
	}
	return requests, nil
}

func (r *Repository) ListOvertimeCandidates(ctx context.Context, businessID int64, date string, thresholdMinutes int32) ([]domain.OvertimeDetectionCandidate, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListOvertimeCandidates(ctx, sqlc.ListOvertimeCandidatesParams{
		DurationMinutes:   thresholdMinutes,
		BusinessID:        businessID,
		Date:              d,
		DurationMinutes_2: thresholdMinutes,
	})
	if err != nil {
		return nil, err
	}
	candidates := make([]domain.OvertimeDetectionCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, domain.OvertimeDetectionCandidate{
			EmployeeID: row.EmployeeID, Date: row.Date.Format("2006-01-02"), ExtraMinutes: row.ExtraMinutes,
		})
	}
	return candidates, nil
}

func overtimeRequestWriteResult(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrOvertimeRequestNotPending
	}
	return nil
}

func (r *Repository) GetOvertimeConfig(ctx context.Context, businessID int64) (*domain.OvertimeConfigRecord, error) {
	row, err := r.queries.GetOvertimeConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOvertimeConfigNotFound
		}
		return nil, err
	}
	cfg := overtimeConfigFromRow(row)
	return &cfg, nil
}

func (r *Repository) UpsertOvertimeConfig(ctx context.Context, businessID int64, cfg domain.OvertimeConfigRecord) error {
	_, err := r.queries.UpsertOvertimeConfig(ctx, sqlc.UpsertOvertimeConfigParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		DailyThresholdMinutes: cfg.DailyThresholdMinutes,
		DefaultMultiplier:     fmt.Sprintf("%.2f", cfg.DefaultMultiplier),
		AutoDetectEnabled:     cfg.AutoDetectEnabled,
	})
	return err
}

func (r *Repository) ListAllConfiguredOwnerIDs(ctx context.Context, limit, offset int32) ([]int64, error) {
	return r.queries.ListAllConfiguredOwnerIDs(ctx, sqlc.ListAllConfiguredOwnerIDsParams{
		Limit:  limit,
		Offset: offset,
	})
}

// --- Helpers ---

func overtimeRequestFromRow(row sqlc.HrOvertimeRequest) domain.OvertimeRequestRecord {
	return domain.OvertimeRequestRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeID:           row.EmployeeID,
		Date:                 row.Date.Format("2006-01-02"),
		PlannedExtraMinutes:  row.PlannedExtraMinutes,
		Reason:               row.Reason.String,
		Status:               row.Status,
		CurrentApprovalLevel: row.CurrentApprovalLevel,
		RejectionReason:      row.RejectionReason.String,
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
		BusinessID:            row.BusinessID,
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

func nullStringFromFloat(f float64) sql.NullString {
	if f == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", f), Valid: true}
}
