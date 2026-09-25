package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.queries.ListAuditLogs(ctx, sqlc.ListAuditLogsParams{
		BusinessID:      filter.BusinessID,
		Action:          nullStringPtr(filter.Action),
		EntityType:      nullStringPtr(filter.EntityType),
		EmployeeID:      nullInt64Ptr(filter.EmployeeID),
		StartDate:       nullTimePtr(filter.StartDate),
		EndDate:         endOfDayTimePtr(filter.EndDate),
		CursorCreatedAt: nullTime(filter.CursorCreatedAt),
		CursorID:        nullInt64Ptr(filter.CursorID),
		Limit:           limit,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list audit logs", logger.Error(err))
		return nil, err
	}

	logs := make([]domain.AuditLog, 0, len(rows))
	for _, row := range rows {
		l := domain.AuditLog{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			EmployeeID: row.EmployeeID,
			Action:     row.Action,
			EntityType: row.EntityType,
			CreatedAt:  row.CreatedAt.Time,
		}
		if row.TerminalID.Valid {
			l.TerminalID = &row.TerminalID.Int64
		}
		if row.EntityID.Valid {
			l.EntityID = &row.EntityID.Int64
		}
		if len(row.Details) > 0 {
			raw := row.Details
			l.Details = &raw
		}
		if row.IpAddress.Valid {
			l.IPAddress = &row.IpAddress.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *Repository) Create(ctx context.Context, input domain.AuditLogInput) (int64, error) {
	details := json.RawMessage("{}")
	if input.Details != nil {
		details = *input.Details
	}

	result, err := r.queries.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		BusinessID: input.BusinessID,
		BranchID:   input.BranchID,
		TerminalID: nullInt64Ptr(input.TerminalID),
		EmployeeID: input.EmployeeID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   nullInt64Ptr(input.EntityID),
		Details:    details,
		IpAddress:  nullStringPtr(input.IPAddress),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create audit log", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func nullStringPtr(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullTimePtr(v *string) sql.NullTime {
	if v == nil || *v == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", *v)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", *v)
		if err != nil {
			return sql.NullTime{}
		}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// endOfDayTimePtr converts a date-only inclusive upper bound to the last
// instant of that day. The query compares created_at <= end_date, so the
// midnight value nullTimePtr returns would drop every row created later on the
// end date (B55-R06). A value that is not date-only keeps its exact instant.
func endOfDayTimePtr(v *string) sql.NullTime {
	if v == nil || *v == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", *v)
	if err != nil {
		return nullTimePtr(v)
	}
	return sql.NullTime{Time: t.Add(24*time.Hour - time.Nanosecond), Valid: true}
}

func nullTime(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *v, Valid: true}
}
