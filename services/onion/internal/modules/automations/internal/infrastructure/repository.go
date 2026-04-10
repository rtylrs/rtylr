package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Automation, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, trigger_type, trigger_config, actions, is_active,
		        last_run_at, run_count, created_at, updated_at
		 FROM crm_automations WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list automations", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Automation, 0)

	for rows.Next() {
		a, scanErr := scanAutomation(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		items = append(items, a)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.AutomationInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_automations
		   (owner_id, name, trigger_type, trigger_config, actions, is_active)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.Name, input.TriggerType,
		string(input.TriggerConfig), string(input.Actions), input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create automation", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.AutomationUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_automations
		 SET name = ?, trigger_type = ?, trigger_config = ?, actions = ?, is_active = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.TriggerType,
		string(input.TriggerConfig), string(input.Actions), input.IsActive,
		input.ID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update automation", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrAutomationNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, automationID int64) error {
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM crm_automations WHERE id = ? AND owner_id = ?",
		automationID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete automation", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrAutomationNotFound
	}

	return nil
}

type automationScanner interface {
	Scan(dest ...interface{}) error
}

func scanAutomation(s automationScanner) (domain.Automation, error) {
	var a domain.Automation
	var triggerConfig, actions sql.NullString
	var lastRunAt sql.NullTime

	if err := s.Scan(
		&a.ID, &a.OwnerID, &a.Name, &a.TriggerType, &triggerConfig, &actions,
		&a.IsActive, &lastRunAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Automation{}, domain.ErrAutomationNotFound
		}

		return domain.Automation{}, err
	}

	if triggerConfig.Valid {
		a.TriggerConfig = []byte(triggerConfig.String)
	} else {
		a.TriggerConfig = []byte("{}")
	}

	if actions.Valid {
		a.Actions = []byte(actions.String)
	} else {
		a.Actions = []byte("[]")
	}

	if lastRunAt.Valid {
		t := lastRunAt.Time
		a.LastRunAt = &t
	}

	return a, nil
}
