package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// projectV2Fields is the shared column shape of the three generated v2 row
// structs (GetProjectV2ByIDRow, ListProjectsV2Row, ListArchivedProjectsV2Row).
// They are field-identical, so one mapper serves all three via this carrier.
type projectV2Row struct {
	ID          int64
	WorkspaceID int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Icon        string
	Color       string
	Status      string
	Visibility  string
	IsPersonal  bool
	StartDate   sql.NullTime
	DueDate     sql.NullTime
	CreatedBy   int64
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	ArchivedAt  sql.NullTime
	PurgeAt     sql.NullTime
	TaskCount   interface{}
	DoneCount   interface{}
	MemberCount interface{}
}

func mapProjectV2(r projectV2Row) domain.Project {
	p := domain.Project{
		ID:             r.ID,
		WorkspaceID:    r.WorkspaceID,
		OwnerID:        r.OwnerID,
		BusinessID:     r.BusinessID,
		Name:           r.Name,
		Description:    r.Description,
		Icon:           r.Icon,
		Color:          r.Color,
		Status:         r.Status,
		Visibility:     r.Visibility,
		IsPersonal:     r.IsPersonal,
		CreatedBy:      r.CreatedBy,
		TaskCount:      int32(countIfaceToInt64(r.TaskCount)),   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CompletedCount: int32(countIfaceToInt64(r.DoneCount)),   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		MemberCount:    int32(countIfaceToInt64(r.MemberCount)), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
	if r.StartDate.Valid {
		p.StartDate = &r.StartDate.Time
	}
	if r.DueDate.Valid {
		p.DueDate = &r.DueDate.Time
	}
	if r.CreatedAt.Valid {
		p.CreatedAt = r.CreatedAt.Time
	}
	if r.UpdatedAt.Valid {
		p.UpdatedAt = r.UpdatedAt.Time
	}
	if r.ArchivedAt.Valid {
		p.ArchivedAt = &r.ArchivedAt.Time
	}
	if r.PurgeAt.Valid {
		p.PurgeAt = &r.PurgeAt.Time
	}
	return p
}

// GetProjectV2ByID returns one business-scoped project with the v2 fields.
func (r *Repository) GetProjectV2ByID(ctx context.Context, id, businessID int64) (domain.Project, error) {
	row, err := r.queries.GetProjectV2ByID(ctx, sqlc.GetProjectV2ByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.Project{}, err
	}
	return mapProjectV2(projectV2Row{
		ID: row.ID, WorkspaceID: row.WorkspaceID, OwnerID: row.OwnerID,
		BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
		Icon: row.Icon, Color: row.Color, Status: string(row.Status), Visibility: string(row.Visibility),
		IsPersonal: row.IsPersonal, StartDate: row.StartDate, DueDate: row.DueDate,
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ArchivedAt: row.ArchivedAt, PurgeAt: row.PurgeAt,
		TaskCount: row.TaskCount, DoneCount: row.CompletedCount, MemberCount: row.MemberCount,
	}), nil
}

// ListProjectsV2 lists active, non-personal projects of one business under the
// curation rule: bypass callers see everything; everyone else sees projects
// they created or belong to.
// ListProjectsV2 lists projects under curation. managedUserIDs widens the scope
// for the dept-manager proxy; an empty or nil slice widens nothing.
// managedInClause shapes the managed-user set for an IN clause. sqlc expands an
// empty slice into an invalid IN (), so an empty set becomes a single 0: no user
// has id 0, which makes IN (0) match nothing without special-casing the SQL.
func managedInClause(ids []int64) []int64 {
	if len(ids) == 0 {
		return []int64{0}
	}
	return ids
}

func (r *Repository) ListProjectsV2(ctx context.Context, businessID, callerID int64, includeAll bool, managedUserIDs []int64, status, visibility, q string, limit, offset int64) ([]domain.Project, int64, error) {
	managed := managedInClause(managedUserIDs)
	all := int64(0)
	if includeAll {
		all = 1
	}
	statusParam := sqlc.NullFlowProjectsStatus{FlowProjectsStatus: sqlc.FlowProjectsStatus(status), Valid: status != ""}
	total, err := r.queries.CountProjectsV2(ctx, sqlc.CountProjectsV2Params{
		BusinessID:      businessID,
		Column2:         all,
		CreatedBy:       callerID,
		UserID:          callerID,
		ManagedCreators: managed,
		ManagedOwners:   managed,
		ManagedMembers:  managed,
		Column8:         status,
		Status:          statusParam,
		Column10:        visibility,
		Visibility:      sqlc.FlowProjectsVisibility(visibility),
		Column12:        q,
		CONCAT:          q,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountProjectsV2 failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListProjectsV2(ctx, sqlc.ListProjectsV2Params{
		BusinessID:      businessID,
		Column2:         all,
		CreatedBy:       callerID,
		UserID:          callerID,
		ManagedCreators: managed,
		ManagedOwners:   managed,
		ManagedMembers:  managed,
		Column8:         status,
		Status:          statusParam,
		Column10:        visibility,
		Visibility:      sqlc.FlowProjectsVisibility(visibility),
		Column12:        q,
		CONCAT:          q,
		Limit:           int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:          int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectsV2 failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, mapProjectV2(projectV2Row{
			ID: row.ID, WorkspaceID: row.WorkspaceID, OwnerID: row.OwnerID,
			BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
			Icon: row.Icon, Color: row.Color, Status: string(row.Status), Visibility: string(row.Visibility),
			IsPersonal: row.IsPersonal, StartDate: row.StartDate, DueDate: row.DueDate,
			CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			ArchivedAt: row.ArchivedAt, PurgeAt: row.PurgeAt,
			TaskCount: row.TaskCount, DoneCount: row.CompletedCount, MemberCount: row.MemberCount,
		}))
	}
	return projects, total, nil
}

// ListArchivedProjectsV2 lists the recycle bin under the same curation rule,
// soonest purge first.
func (r *Repository) ListArchivedProjectsV2(ctx context.Context, businessID, callerID int64, includeAll bool, limit, offset int64) ([]domain.Project, int64, error) {
	all := int64(0)
	if includeAll {
		all = 1
	}
	total, err := r.queries.CountArchivedProjectsV2(ctx, sqlc.CountArchivedProjectsV2Params{
		BusinessID: businessID,
		Column2:    all,
		CreatedBy:  callerID,
		UserID:     callerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountArchivedProjectsV2 failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListArchivedProjectsV2(ctx, sqlc.ListArchivedProjectsV2Params{
		BusinessID: businessID,
		Column2:    all,
		CreatedBy:  callerID,
		UserID:     callerID,
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListArchivedProjectsV2 failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, mapProjectV2(projectV2Row{
			ID: row.ID, WorkspaceID: row.WorkspaceID, OwnerID: row.OwnerID,
			BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
			Icon: row.Icon, Color: row.Color, Status: string(row.Status), Visibility: string(row.Visibility),
			IsPersonal: row.IsPersonal, StartDate: row.StartDate, DueDate: row.DueDate,
			CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			ArchivedAt: row.ArchivedAt, PurgeAt: row.PurgeAt,
			TaskCount: row.TaskCount, DoneCount: row.CompletedCount, MemberCount: row.MemberCount,
		}))
	}
	return projects, total, nil
}

// UpdateProjectVisibility flips private/public on one business project.
func (r *Repository) UpdateProjectVisibility(ctx context.Context, id, businessID int64, visibility string) error {
	_, err := r.queries.UpdateProjectVisibility(ctx, sqlc.UpdateProjectVisibilityParams{
		Visibility: sqlc.FlowProjectsVisibility(visibility),
		ID:         id,
		BusinessID: businessID,
	})
	return err
}

// RestoreProjectV2 pulls a binned project back to active and clears the
// archive timestamps.
func (r *Repository) RestoreProjectV2(ctx context.Context, id, businessID int64) error {
	_, err := r.queries.RestoreProjectV2(ctx, sqlc.RestoreProjectV2Params{ID: id, BusinessID: businessID})
	return err
}

// ArchiveProject stamps the recycle-bin timestamps on archive. purgeAt is
// computed by the service from the business's retention setting; nil keeps the
// legacy v1 behaviour of an open-ended archive while the retention config
// plumbing is still landing.
func (r *Repository) ArchiveProject(ctx context.Context, id, businessID int64, purgeAt *time.Time) error {
	err := r.queries.ArchiveProject(ctx, sqlc.ArchiveProjectParams{
		PurgeAt:    nullTimePtrToSql(purgeAt),
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ArchiveProject failed", logger.Error(err))
		return err
	}
	return nil
}

// PurgeProjectCascade irreversibly deletes one archived project and every row
// that hangs off it, in a single transaction. Attachment object URLs are read
// before deletion and returned so the caller can best-effort sweep the stored
// files after the rows are gone.
func (r *Repository) PurgeProjectCascade(ctx context.Context, id, businessID int64) ([]string, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var urls []string
	rows, err := tx.QueryContext(ctx,
		`SELECT a.url FROM flow_task_attachments a
		 JOIN flow_tasks t ON t.id = a.task_id
		 WHERE t.project_id = ? AND t.business_id = ?`, id, businessID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("scan attachment URL: %w; close rows: %w", err, closeErr)
			}
			return nil, err
		}
		urls = append(urls, u)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stmts := []string{
		`DELETE a FROM flow_task_attachments a JOIN flow_tasks t ON t.id = a.task_id WHERE t.project_id = ? AND t.business_id = ?`,
		`DELETE tl FROM flow_time_logs tl JOIN flow_tasks t ON t.id = tl.task_id WHERE t.project_id = ? AND t.business_id = ?`,
		`DELETE c FROM flow_task_comments c JOIN flow_tasks t ON t.id = c.task_id WHERE t.project_id = ? AND t.business_id = ?`,
		`DELETE act FROM flow_task_activity act JOIN flow_tasks t ON t.id = act.task_id WHERE t.project_id = ? AND t.business_id = ?`,
		`DELETE fp FROM flow_task_focus_pins fp JOIN flow_tasks t ON t.id = fp.task_id WHERE t.project_id = ? AND t.business_id = ?`,
		`DELETE FROM flow_task_dependencies WHERE task_id IN (SELECT id FROM flow_tasks WHERE project_id = ? AND business_id = ?) OR depends_on_task_id IN (SELECT id FROM flow_tasks WHERE project_id = ? AND business_id = ?)`,
		`DELETE ct FROM flow_cycle_tasks ct JOIN flow_cycles cy ON cy.id = ct.cycle_id WHERE cy.project_id = ? AND cy.business_id = ?`,
		`DELETE mt FROM flow_milestone_tasks mt JOIN flow_milestones m ON m.id = mt.milestone_id WHERE m.project_id = ? AND m.business_id = ?`,
		`UPDATE flow_tasks SET parent_task_id = NULL WHERE parent_task_id IN (SELECT id FROM (SELECT id FROM flow_tasks WHERE project_id = ? AND business_id = ?) x)`,
		`DELETE FROM flow_tasks WHERE project_id = ? AND business_id = ?`,
		`DELETE FROM flow_project_columns WHERE project_id = ?`,
		`DELETE FROM flow_milestones WHERE project_id = ? AND business_id = ?`,
		`DELETE FROM flow_cycles WHERE project_id = ? AND business_id = ?`,
		`DELETE FROM flow_projects WHERE id = ? AND business_id = ?`,
	}
	for _, stmt := range stmts {
		var args []interface{}
		switch strings.Count(stmt, "?") {
		case 1:
			args = []interface{}{id}
		case 4:
			args = []interface{}{id, businessID, id, businessID}
		default:
			args = []interface{}{id, businessID}
		}
		if _, err := tx.ExecContext(ctx, stmt, args...); err != nil {
			logger.Ctx(ctx).Error("PurgeProjectCascade statement failed", logger.String("stmt", stmt[:40]), logger.Error(err))
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return urls, nil
}

// ListPublicProjectsByOwnerV2 returns the subject's public, non-archived
// projects inside one business, metadata only.
func (r *Repository) ListPublicProjectsByOwnerV2(
	ctx context.Context,
	businessID, ownerID int64,
	limit int32,
) ([]domain.ProfileProject, error) {
	rows, err := r.queries.ListPublicProjectsByOwnerV2(ctx, sqlc.ListPublicProjectsByOwnerV2Params{
		BusinessID: businessID,
		CreatedBy:  ownerID,
		Limit:      limit,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListPublicProjectsByOwnerV2 failed", logger.Error(err))
		return nil, err
	}
	out := make([]domain.ProfileProject, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ProfileProject{
			ID:             row.ID,
			Name:           row.Name,
			Icon:           row.Icon,
			Color:          row.Color,
			Description:    row.Description,
			Status:         string(row.Status),
			TaskCount:      int32(row.TaskCount),      // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CompletedCount: int32(row.CompletedCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	return out, nil
}
