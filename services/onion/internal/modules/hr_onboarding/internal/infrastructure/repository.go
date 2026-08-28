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

func (r *Repository) CreateTemplate(ctx context.Context, businessID int64, params domain.CreateTemplateParams) (*domain.ChecklistTemplateRecord, error) {
	result, err := r.queries.CreateChecklistTemplate(ctx, sqlc.CreateChecklistTemplateParams{
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

	return r.GetTemplate(ctx, businessID, id)
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

	items, err := r.queries.ListChecklistItems(ctx, sqlc.ListChecklistItemsParams{
		BusinessID: businessID,
		TemplateID: id,
	})
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

func (r *Repository) ListTemplates(ctx context.Context, businessID int64, limit, offset int32) ([]domain.ChecklistTemplateRecord, int64, error) {
	total, err := r.queries.CountChecklistTemplates(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListChecklistTemplates(ctx, sqlc.ListChecklistTemplatesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	templates := make([]domain.ChecklistTemplateRecord, 0, len(rows))
	for _, row := range rows {
		templates = append(templates, templateFromRow(row))
	}

	return templates, total, nil
}

func (r *Repository) UpdateTemplate(ctx context.Context, businessID, id int64, params domain.UpdateTemplateParams) error {
	return r.queries.UpdateChecklistTemplate(ctx, sqlc.UpdateChecklistTemplateParams{
		Name:        params.Name,
		TriggerType: params.TriggerType,
		RoleID:      sqlutil.NullInt64(params.RoleID),
		BusinessID:  businessID,
		ID:          id,
	})
}

func (r *Repository) CreateTemplateItem(ctx context.Context, businessID, templateID int64, params domain.CreateTemplateItemParams) (*domain.ChecklistItemRecord, error) {
	result, err := r.queries.CreateChecklistItem(ctx, sqlc.CreateChecklistItemParams{
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

	row, err := r.queries.GetChecklistItem(ctx, sqlc.GetChecklistItemParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		return nil, err
	}

	item := checklistItemFromRow(row)
	return &item, nil
}

func (r *Repository) ListTemplateItems(ctx context.Context, businessID, templateID int64) ([]domain.ChecklistItemRecord, error) {
	rows, err := r.queries.ListChecklistItems(ctx, sqlc.ListChecklistItemsParams{
		BusinessID: businessID,
		TemplateID: templateID,
	})
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
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
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
		_ = tx.Rollback()
		return nil, err
	}

	checklistID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	templateItems, err := qtx.ListChecklistItems(ctx, sqlc.ListChecklistItemsParams{
		BusinessID: businessID,
		TemplateID: templateID,
	})
	if err != nil {
		_ = tx.Rollback()
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
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetEmployeeChecklist(ctx, businessID, employeeID, checklistID)
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

	items, err := r.queries.ListEmployeeChecklistItems(ctx, sqlc.ListEmployeeChecklistItemsParams{
		BusinessID:          businessID,
		EmployeeChecklistID: checklistID,
	})
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

func (r *Repository) ListEmployeeChecklists(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.EmployeeChecklistRecord, int64, error) {
	total, err := r.queries.CountEmployeeChecklists(ctx, sqlc.CountEmployeeChecklistsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListEmployeeChecklists(ctx, sqlc.ListEmployeeChecklistsParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	checklists := make([]domain.EmployeeChecklistRecord, 0, len(rows))
	for _, row := range rows {
		checklists = append(checklists, employeeChecklistFromRow(row))
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
	return r.queries.CompleteEmployeeChecklistItem(ctx, sqlc.CompleteEmployeeChecklistItemParams{
		CompletedBy: sql.NullInt64{Int64: completedBy, Valid: true},
		BusinessID:  businessID,
		ID:          itemID,
	})
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
	rows, err := r.queries.ListEmployeeChecklistItems(ctx, sqlc.ListEmployeeChecklistItemsParams{
		BusinessID:          businessID,
		EmployeeChecklistID: employeeChecklistID,
	})
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
