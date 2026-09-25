package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure/persistence/sqlc"
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

func nullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

// The generated item queries intentionally have a LIMIT 500, but the public
// records expose the complete nested item collection. Keep the unbounded
// reads in the repository so list and detail responses cannot silently drop
// tasks after the first page of items. The parent queries remain paginated.
const (
	listOnboardingChecklistItemsQuery = `
		SELECT id, owner_id, business_id, template_id, title, description,
		       assignee_role, due_offset_days, requires_attachment,
		       created_at, updated_at
		FROM hr_onboarding_checklist_items
		WHERE business_id = ? AND template_id = ?
		ORDER BY due_offset_days ASC, id ASC`

	listOnboardingEmployeeChecklistItemsQuery = `
		SELECT id, owner_id, business_id, employee_checklist_id, checklist_item_id,
		       assigned_to_employee_id, due_date, status, completed_at, completed_by,
		       attachment_url, created_at, updated_at
		FROM hr_onboarding_employee_checklist_items
		WHERE business_id = ? AND employee_checklist_id = ?
		ORDER BY due_date ASC, id ASC`

	listOnboardingEmployeeChecklistsQuery = `
		SELECT id, owner_id, business_id, employee_id, template_id, status,
		       triggered_at, created_at, updated_at
		FROM hr_onboarding_employee_checklists
		WHERE business_id = ? AND employee_id = ?
		ORDER BY triggered_at DESC, id DESC
		LIMIT ? OFFSET ?`
)

func listChecklistItems(ctx context.Context, q sqlc.DBTX, businessID, templateID int64) ([]sqlc.HrOnboardingChecklistItem, error) {
	rows, err := q.QueryContext(ctx, listOnboardingChecklistItemsQuery, businessID, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]sqlc.HrOnboardingChecklistItem, 0)
	for rows.Next() {
		var item sqlc.HrOnboardingChecklistItem
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.BusinessID,
			&item.TemplateID,
			&item.Title,
			&item.Description,
			&item.AssigneeRole,
			&item.DueOffsetDays,
			&item.RequiresAttachment,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func listEmployeeChecklistItems(ctx context.Context, q sqlc.DBTX, businessID, employeeChecklistID int64) ([]sqlc.HrOnboardingEmployeeChecklistItem, error) {
	rows, err := q.QueryContext(ctx, listOnboardingEmployeeChecklistItemsQuery, businessID, employeeChecklistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]sqlc.HrOnboardingEmployeeChecklistItem, 0)
	for rows.Next() {
		var item sqlc.HrOnboardingEmployeeChecklistItem
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.BusinessID,
			&item.EmployeeChecklistID,
			&item.ChecklistItemID,
			&item.AssignedToEmployeeID,
			&item.DueDate,
			&item.Status,
			&item.CompletedAt,
			&item.CompletedBy,
			&item.AttachmentUrl,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func listEmployeeChecklistRows(ctx context.Context, q sqlc.DBTX, businessID, employeeID int64, limit, offset int32) ([]sqlc.HrOnboardingEmployeeChecklist, error) {
	rows, err := q.QueryContext(ctx, listOnboardingEmployeeChecklistsQuery, businessID, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checklists := make([]sqlc.HrOnboardingEmployeeChecklist, 0)
	for rows.Next() {
		var checklist sqlc.HrOnboardingEmployeeChecklist
		if err := rows.Scan(
			&checklist.ID,
			&checklist.OwnerID,
			&checklist.BusinessID,
			&checklist.EmployeeID,
			&checklist.TemplateID,
			&checklist.Status,
			&checklist.TriggeredAt,
			&checklist.CreatedAt,
			&checklist.UpdatedAt,
		); err != nil {
			return nil, err
		}
		checklists = append(checklists, checklist)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return checklists, nil
}

func templateWithItems(ctx context.Context, q sqlc.DBTX, businessID int64, row sqlc.HrOnboardingChecklistTemplate) (*domain.ChecklistTemplateRecord, error) {
	items, err := listChecklistItems(ctx, q, businessID, row.ID)
	if err != nil {
		return nil, err
	}
	record := templateFromRow(row)
	record.Items = make([]domain.ChecklistItemRecord, 0, len(items))
	for _, item := range items {
		record.Items = append(record.Items, checklistItemFromRow(item))
	}
	return &record, nil
}

func employeeChecklistWithItemsFrom(ctx context.Context, q sqlc.DBTX, businessID int64, row sqlc.HrOnboardingEmployeeChecklist) (*domain.EmployeeChecklistRecord, error) {
	items, err := listEmployeeChecklistItems(ctx, q, businessID, row.ID)
	if err != nil {
		return nil, err
	}
	record := employeeChecklistFromRow(row)
	record.Items = make([]domain.EmployeeChecklistItemRecord, 0, len(items))
	for _, item := range items {
		record.Items = append(record.Items, employeeChecklistItemFromRow(item))
	}
	return &record, nil
}

func (r *Repository) CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
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
	result, err := qtx.CreateChecklistTemplate(ctx, sqlc.CreateChecklistTemplateParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        params.Name,
		TriggerType: params.TriggerType,
		RoleID:      sqlutil.NullInt64(params.RoleID),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	row, err := qtx.GetChecklistTemplate(ctx, sqlc.GetChecklistTemplateParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		return nil, err
	}
	record, err := templateWithItems(ctx, tx, businessID, row)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return record, nil
}
