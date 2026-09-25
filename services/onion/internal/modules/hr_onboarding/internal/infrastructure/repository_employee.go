package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) GetEmployeeChecklist(ctx context.Context, businessID, employeeID, checklistID int64) (*domain.EmployeeChecklistRecord, error) {
	row, err := r.queries.GetEmployeeChecklist(ctx, sqlc.GetEmployeeChecklistParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		ID:         checklistID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrChecklistNotFound
		}
		return nil, err
	}

	return r.employeeChecklistWithItems(ctx, businessID, row)
}

// GetLatestEmployeeChecklist returns the employee's most recently triggered
// checklist. The public GET /employees/{employee_id} route carries no checklist
// id, so resolving the employee's latest checklist is the only way it can
// return real data (the employee id must never be used as a checklist id).
func (r *Repository) GetLatestEmployeeChecklist(ctx context.Context, businessID, employeeID int64) (*domain.EmployeeChecklistRecord, error) {
	row, err := r.queries.GetLatestEmployeeChecklist(ctx, sqlc.GetLatestEmployeeChecklistParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrChecklistNotFound
		}
		return nil, err
	}

	return r.employeeChecklistWithItems(ctx, businessID, row)
}

func (r *Repository) employeeChecklistWithItems(ctx context.Context, businessID int64, row sqlc.HrOnboardingEmployeeChecklist) (*domain.EmployeeChecklistRecord, error) {
	return employeeChecklistWithItemsFrom(ctx, r.pool, businessID, row)
}

func (r *Repository) ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeChecklistRecord, int64, error) {
	total, err := r.queries.CountEmployeeChecklists(ctx, sqlc.CountEmployeeChecklistsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := listEmployeeChecklistRows(ctx, r.pool, businessID, employeeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	checklists := make([]domain.EmployeeChecklistRecord, 0, len(rows))
	for _, row := range rows {
		record, err := employeeChecklistWithItemsFrom(ctx, r.pool, businessID, row)
		if err != nil {
			return nil, 0, err
		}
		checklists = append(checklists, *record)
	}

	return checklists, total, nil
}

func (r *Repository) UpdateEmployeeChecklistStatus(ctx context.Context, businessID, checklistID int64, status string) error {
	return r.queries.UpdateEmployeeChecklistStatus(ctx, sqlc.UpdateEmployeeChecklistStatusParams{
		Status:     status,
		BusinessID: businessID,
		ID:         checklistID,
	})
}

func (r *Repository) CreateEmployeeChecklistItem(ctx context.Context, businessID, employeeChecklistID, checklistItemID, assignedToEmployeeID int64, dueDate string) error {
	parsed, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return fmt.Errorf("invalid due_date format: %w", err)
	}

	return r.queries.CreateEmployeeChecklistItem(ctx, sqlc.CreateEmployeeChecklistItemParams{
		OwnerID:              businessID,
		BusinessID:           businessID,
		EmployeeChecklistID:  employeeChecklistID,
		ChecklistItemID:      checklistItemID,
		AssignedToEmployeeID: assignedToEmployeeID,
		DueDate:              parsed,
		Status:               "incomplete",
	})
}

func (r *Repository) GetChecklistItem(ctx context.Context, businessID, itemID int64) (*domain.EmployeeChecklistItemRecord, error) {
	row, err := r.queries.GetEmployeeChecklistItem(ctx, sqlc.GetEmployeeChecklistItemParams{
		BusinessID: businessID,
		ID:         itemID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrChecklistItemNotFound
		}
		return nil, err
	}

	item := employeeChecklistItemFromRow(row)
	return &item, nil
}

func (r *Repository) CompleteChecklistItem(ctx context.Context, businessID, itemID, completedBy int64) error {
	return r.CompleteChecklistItemWithAttachment(ctx, businessID, itemID, completedBy, nil)
}

func (r *Repository) CompleteChecklistItemWithAttachment(ctx context.Context, businessID, itemID, completedBy int64, attachmentURL *string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	item, err := qtx.GetEmployeeChecklistItem(ctx, sqlc.GetEmployeeChecklistItemParams{
		BusinessID: businessID,
		ID:         itemID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrChecklistItemNotFound
		}
		return err
	}

	result, err := qtx.CompleteEmployeeChecklistItem(ctx, sqlc.CompleteEmployeeChecklistItemParams{
		CompletedBy:   sql.NullInt64{Int64: completedBy, Valid: true},
		AttachmentUrl: nullString(attachmentURL),
		BusinessID:    businessID,
		ID:            itemID,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// The item was read above, so a zero-row conditional update means a
		// concurrent request won the race (or the item was completed between
		// the read and the update). Treat both cases as the same conflict.
		return domain.ErrChecklistAlreadyCompleted
	}

	total, err := qtx.CountChecklistItemsTotal(ctx, sqlc.CountChecklistItemsTotalParams{
		BusinessID:          businessID,
		EmployeeChecklistID: item.EmployeeChecklistID,
	})
	if err != nil {
		return err
	}
	completed, err := qtx.CountChecklistItemsCompleted(ctx, sqlc.CountChecklistItemsCompletedParams{
		BusinessID:          businessID,
		EmployeeChecklistID: item.EmployeeChecklistID,
	})
	if err != nil {
		return err
	}
	if total > 0 && total == completed {
		if err := qtx.UpdateEmployeeChecklistStatus(ctx, sqlc.UpdateEmployeeChecklistStatusParams{
			Status:     domain.ChecklistStatusCompleted,
			BusinessID: businessID,
			ID:         item.EmployeeChecklistID,
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *Repository) EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error) {
	var exists int
	err := r.pool.QueryRowContext(ctx,
		`SELECT 1
FROM hr_employees
WHERE business_id = ? AND id = ? AND status <> 'terminated'
LIMIT 1`, businessID, employeeID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repository) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	var employeeID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT id
FROM hr_employees
WHERE business_id = ? AND user_id = ? AND status <> 'terminated'
ORDER BY id ASC
LIMIT 1`, businessID, userID,
	).Scan(&employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return 0, err
	}
	return employeeID, nil
}

func (r *Repository) GetChecklistCompletionStats(ctx context.Context, businessID, checklistID int64) (int, int, error) {
	total, err := r.queries.CountChecklistItemsTotal(ctx, sqlc.CountChecklistItemsTotalParams{
		BusinessID:          businessID,
		EmployeeChecklistID: checklistID,
	})
	if err != nil {
		return 0, 0, err
	}

	completed, err := r.queries.CountChecklistItemsCompleted(ctx, sqlc.CountChecklistItemsCompletedParams{
		BusinessID:          businessID,
		EmployeeChecklistID: checklistID,
	})
	if err != nil {
		return 0, 0, err
	}

	return int(total), int(completed), nil
}

func (r *Repository) ListEmployeeChecklistItems(ctx context.Context, businessID, employeeChecklistID int64) ([]domain.EmployeeChecklistItemRecord, error) {
	rows, err := listEmployeeChecklistItems(ctx, r.pool, businessID, employeeChecklistID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.EmployeeChecklistItemRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, employeeChecklistItemFromRow(row))
	}

	return items, nil
}

func templateFromRow(row sqlc.HrOnboardingChecklistTemplate) domain.ChecklistTemplateRecord {
	return domain.ChecklistTemplateRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		TriggerType: row.TriggerType,
		RoleID:      row.RoleID.Int64,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func checklistItemFromRow(row sqlc.HrOnboardingChecklistItem) domain.ChecklistItemRecord {
	return domain.ChecklistItemRecord{
		ID:                 row.ID,
		OwnerID:            row.OwnerID,
		BusinessID:         row.BusinessID,
		TemplateID:         row.TemplateID,
		Title:              row.Title,
		Description:        row.Description.String,
		AssigneeRole:       row.AssigneeRole,
		DueOffsetDays:      row.DueOffsetDays,
		RequiresAttachment: row.RequiresAttachment,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func employeeChecklistFromRow(row sqlc.HrOnboardingEmployeeChecklist) domain.EmployeeChecklistRecord {
	return domain.EmployeeChecklistRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		EmployeeID:  row.EmployeeID,
		TemplateID:  row.TemplateID,
		Status:      row.Status,
		TriggeredAt: row.TriggeredAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func employeeChecklistItemFromRow(row sqlc.HrOnboardingEmployeeChecklistItem) domain.EmployeeChecklistItemRecord {
	return domain.EmployeeChecklistItemRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeChecklistID:  row.EmployeeChecklistID,
		ChecklistItemID:      row.ChecklistItemID,
		AssignedToEmployeeID: row.AssignedToEmployeeID,
		DueDate:              row.DueDate.Format("2006-01-02"),
		Status:               row.Status,
		CompletedAt:          row.CompletedAt.Time,
		CompletedBy:          row.CompletedBy.Int64,
		AttachmentURL:        row.AttachmentUrl.String,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}
