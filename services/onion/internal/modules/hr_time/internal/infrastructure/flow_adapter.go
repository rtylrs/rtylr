package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type FlowAdapter struct {
	pool *sql.DB
}

func NewFlowAdapter() *FlowAdapter {
	return &FlowAdapter{pool: dbclient.Pool()}
}

func (a *FlowAdapter) GetCompletedTasksForEmployee(ctx context.Context, ownerID, employeeID int64, periodStart, periodEnd string) ([]domain.FlowTaskRecord, error) {
	ps, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, err
	}
	pe, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, err
	}

	var userID sql.NullInt64
	err = a.pool.QueryRowContext(ctx,
		"SELECT user_id FROM hr_employees WHERE id = ? AND owner_id = ?",
		employeeID, ownerID,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Ctx(ctx).Error("flow adapter: failed to resolve employee user_id",
			logger.Any("owner_id", ownerID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, err
	}
	if !userID.Valid {
		return nil, nil
	}

	// cross-schema read from flow module — completed tasks and logged time for employee import
	rows, err := a.pool.QueryContext(ctx, `
		SELECT t.id, t.title,
		       COALESCE(SUM(tl.duration_minutes), 0) AS total_minutes,
		       t.completed_at
		FROM flow_tasks t
		LEFT JOIN flow_time_logs tl ON tl.task_id = t.id
		WHERE t.assignee_id = ? AND t.owner_id = ? AND t.status = 'done'
		      AND t.completed_at >= ? AND t.completed_at <= ?
		GROUP BY t.id, t.title, t.completed_at
	`, userID.Int64, ownerID, ps, pe)
	if err != nil {
		logger.Ctx(ctx).Error("flow adapter: failed to query completed tasks",
			logger.Any("owner_id", ownerID), logger.Any("user_id", userID.Int64), logger.Error(err))
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
