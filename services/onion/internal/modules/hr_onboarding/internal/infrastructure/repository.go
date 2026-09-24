package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
