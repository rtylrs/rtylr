package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) GetTemplate(ctx context.Context, businessID, id int64) (*domain.ChecklistTemplateRecord, error) {
	row, err := r.queries.GetChecklistTemplate(ctx, sqlc.GetChecklistTemplateParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, err
	}

	record, err := templateWithItems(ctx, r.pool, businessID, row)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *Repository) ListTemplates(ctx context.Context, businessID int64, limit, offset int32) ([]domain.ChecklistTemplateRecord, int64, error) {
	return r.ListTemplatesFiltered(ctx, businessID, domain.ListTemplatesParams{Limit: limit, Offset: offset})
}

func (r *Repository) ListTemplatesFiltered(ctx context.Context, businessID int64, params domain.ListTemplatesParams) ([]domain.ChecklistTemplateRecord, int64, error) {
	triggerType := nullString(params.TriggerType)
	total, err := r.queries.CountChecklistTemplates(ctx, sqlc.CountChecklistTemplatesParams{
		BusinessID:  businessID,
		TriggerType: triggerType,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListChecklistTemplates(ctx, sqlc.ListChecklistTemplatesParams{
		BusinessID:  businessID,
		TriggerType: triggerType,
		Limit:       params.Limit,
		Offset:      params.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	templates := make([]domain.ChecklistTemplateRecord, 0, len(rows))
	for _, row := range rows {
		record, err := templateWithItems(ctx, r.pool, businessID, row)
		if err != nil {
			return nil, 0, err
		}
		templates = append(templates, *record)
	}

	return templates, total, nil
}

func (r *Repository) UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error {
	result, err := r.queries.UpdateChecklistTemplate(ctx, sqlc.UpdateChecklistTemplateParams{
		Name:        params.Name,
		TriggerType: nullString(params.TriggerType),
		RoleID:      nullInt64(params.RoleID),
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		// MySQL reports changed rows, so an idempotent update can affect zero
		// rows even when the record exists. Probe the tenant-scoped key before
		// deciding that this was a missing template.
		var exists int
		probeErr := r.pool.QueryRowContext(ctx,
			`SELECT 1 FROM hr_onboarding_checklist_templates WHERE business_id = ? AND id = ?`,
			businessID, id,
		).Scan(&exists)
		if errors.Is(probeErr, sql.ErrNoRows) {
			return domain.ErrTemplateNotFound
		}
		if probeErr != nil {
			return probeErr
		}
	}
	return nil
}

func (r *Repository) CreateTemplateItem(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateChecklistItem(ctx, sqlc.CreateChecklistItemParams{
		OwnerID:            businessID,
		BusinessID:         businessID,
		TemplateID:         templateID,
		Title:              params.Title,
		Description:        sqlutil.NullString(params.Description),
		AssigneeRole:       params.AssigneeRole,
		DueOffsetDays:      params.DueOffsetDays,
		RequiresAttachment: params.RequiresAttachment,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	row, err := qtx.GetChecklistItem(ctx, sqlc.GetChecklistItemParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		return nil, err
	}

	item := checklistItemFromRow(row)
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return &item, nil
}

func (r *Repository) ListTemplateItems(ctx context.Context, businessID, templateID int64) ([]domain.ChecklistItemRecord, error) {
	rows, err := listChecklistItems(ctx, r.pool, businessID, templateID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.ChecklistItemRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, checklistItemFromRow(row))
	}

	return items, nil
}

func (r *Repository) TriggerChecklist(ctx context.Context, businessID, employeeID, templateID int64, triggerDate time.Time) (*domain.EmployeeChecklistRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	result, err := qtx.CreateEmployeeChecklist(ctx, sqlc.CreateEmployeeChecklistParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		EmployeeID:  employeeID,
		TemplateID:  templateID,
		Status:      "in_progress",
		TriggeredAt: triggerDate,
	})
	if err != nil {
		return nil, err
	}

	checklistID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	templateItems, err := listChecklistItems(ctx, tx, businessID, templateID)
	if err != nil {
		return nil, err
	}

	for _, item := range templateItems {
		dueDate := triggerDate.AddDate(0, 0, int(item.DueOffsetDays))
		err := qtx.CreateEmployeeChecklistItem(ctx, sqlc.CreateEmployeeChecklistItemParams{
			OwnerID:              businessID,
			BusinessID:           businessID,
			EmployeeChecklistID:  checklistID,
			ChecklistItemID:      item.ID,
			AssignedToEmployeeID: employeeID,
			DueDate:              dueDate,
			Status:               "incomplete",
		})
		if err != nil {
			return nil, err
		}
	}

	// Assemble the response while the transaction is still open. Returning a
	// post-commit read would make a transient read failure look like a failed
	// trigger even though the checklist and all of its items are durable.
	row, err := qtx.GetEmployeeChecklist(ctx, sqlc.GetEmployeeChecklistParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		ID:         checklistID,
	})
	if err != nil {
		return nil, err
	}
	record, err := employeeChecklistWithItemsFrom(ctx, tx, businessID, row)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return record, nil
}
