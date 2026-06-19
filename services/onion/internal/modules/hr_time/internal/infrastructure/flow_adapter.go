package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type FlowAdapter struct {
	pool *sql.DB
}

func NewFlowAdapter(db *sql.DB) *FlowAdapter {
	return &FlowAdapter{pool: db}
}

func (a *FlowAdapter) GetCompletedTasksForEmployee(ctx context.Context, businessID, employeeID int64, periodStart, periodEnd string) ([]domain.FlowTaskRecord, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}

	// cross-schema read from hr_core module — resolve employee user_id for flow task lookup
	userIDSQL, userIDArgs, err := sq.Select("user_id").
		From("hr_employees").
		Where(sq.Eq{"id": employeeID, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("flow adapter: failed to build employee user_id query",
			logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, err
	}

	var userID sql.NullInt64
	err = a.pool.QueryRowContext(ctx, userIDSQL, userIDArgs...).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Ctx(ctx).Error("flow adapter: failed to resolve employee user_id",
			logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, err
	}
	if !userID.Valid {
		return nil, nil
	}

	// cross-schema read from flow module — completed tasks and logged time for employee import
	tasksSQL, tasksArgs, err := sq.Select(
		"t.id",
		"t.title",
		"COALESCE(SUM(tl.duration_minutes), 0) AS total_minutes",
		"t.completed_at",
	).
		From("flow_tasks t").
		LeftJoin("flow_time_logs tl ON tl.task_id = t.id").
		Where(sq.Eq{"t.assignee_id": userID.Int64, "t.business_id": businessID, "t.status": "done"}).
		Where(sq.GtOrEq{"t.completed_at": ps}).
		Where(sq.LtOrEq{"t.completed_at": pe}).
		GroupBy("t.id", "t.title", "t.completed_at").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("flow adapter: failed to build completed tasks query",
			logger.Any("business_id", businessID), logger.Any("user_id", userID.Int64), logger.Error(err))
		return nil, err
	}

	rows, err := a.pool.QueryContext(ctx, tasksSQL, tasksArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("flow adapter: failed to query completed tasks",
			logger.Any("business_id", businessID), logger.Any("user_id", userID.Int64), logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("flow adapter: failed to close rows", logger.Error(closeErr))
		}
	}()

	var tasks []domain.FlowTaskRecord
	for rows.Next() {
		var t domain.FlowTaskRecord
		var completedAt sql.NullTime
		if err := rows.Scan(&t.TaskID, &t.Title, &t.DurationMinutes, &completedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.Time.Format("2006-01-02")
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
