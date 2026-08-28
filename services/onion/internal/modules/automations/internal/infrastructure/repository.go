package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Automation, error) {
	rows, err := r.queries.ListAutomations(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list automations", logger.Error(err))

		return nil, err
	}

	items := make([]domain.Automation, 0, len(rows))

	for _, row := range rows {
		items = append(items, automationFromRow(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, input domain.AutomationInput) (int64, error) {
	result, err := r.queries.CreateAutomation(ctx, sqlc.CreateAutomationParams{
		OwnerID:       input.OwnerID,
		BusinessID:    input.BusinessID,
		Name:          input.Name,
		TriggerType:   input.TriggerType,
		TriggerConfig: json.RawMessage(input.TriggerConfig),
		Actions:       json.RawMessage(input.Actions),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create automation", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.AutomationUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	result, err := r.queries.UpdateAutomation(ctx, sqlc.UpdateAutomationParams{
		Name:          input.Name,
		TriggerType:   input.TriggerType,
		TriggerConfig: json.RawMessage(input.TriggerConfig),
		Actions:       json.RawMessage(input.Actions),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:            input.ID,
		BusinessID:    businessID,
	})
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

func (r *Repository) Delete(ctx context.Context, businessID, automationID int64) error {
	result, err := r.queries.DeleteAutomation(ctx, sqlc.DeleteAutomationParams{
		ID:         automationID,
		BusinessID: businessID,
	})
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

func automationFromRow(row sqlc.CrmAutomation) domain.Automation {
	a := domain.Automation{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		TriggerType: row.TriggerType,
		IsActive:    row.IsActive.Bool,
		RunCount:    int(row.RunCount.Int32),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}

	if row.TriggerConfig != nil {
		a.TriggerConfig = []byte(row.TriggerConfig)
	} else {
		a.TriggerConfig = []byte("{}")
	}

	if row.Actions != nil {
		a.Actions = []byte(row.Actions)
	} else {
		a.Actions = []byte("[]")
	}

	if row.LastRunAt.Valid {
		t := row.LastRunAt.Time
		a.LastRunAt = &t
	}

	return a
}
