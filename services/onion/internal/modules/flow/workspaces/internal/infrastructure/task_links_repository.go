package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
)

func mapTaskLink(row sqlc.FlowTaskLink) domain.TaskLink {
	return domain.TaskLink{
		ID: row.ID, TaskID: row.TaskID, BusinessID: row.BusinessID, URL: row.Url, Kind: row.Kind,
		Title: row.Title, Ref: row.Ref, State: row.State, AddedBy: row.AddedBy, CreatedAt: row.CreatedAt,
	}
}

func (r *Repository) CreateTaskLink(ctx context.Context, link domain.TaskLink) (int64, error) {
	res, err := r.queries.CreateTaskLink(ctx, sqlc.CreateTaskLinkParams{
		TaskID: link.TaskID, BusinessID: link.BusinessID, Url: link.URL, Kind: link.Kind,
		Title: link.Title, Ref: link.Ref, State: link.State, AddedBy: link.AddedBy,
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetTaskLink(ctx context.Context, id, businessID int64) (domain.TaskLink, error) {
	row, err := r.queries.GetTaskLink(ctx, sqlc.GetTaskLinkParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.TaskLink{}, err
	}
	return mapTaskLink(row), nil
}

// FindTaskLinkByURL returns found=false, not an error, when the task lacks the URL.
func (r *Repository) FindTaskLinkByURL(ctx context.Context, taskID, businessID int64, url string) (domain.TaskLink, bool, error) {
	row, err := r.queries.FindTaskLinkByURL(ctx, sqlc.FindTaskLinkByURLParams{TaskID: taskID, BusinessID: businessID, Url: url})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TaskLink{}, false, nil
	}
	if err != nil {
		return domain.TaskLink{}, false, err
	}
	return mapTaskLink(row), true, nil
}

func (r *Repository) ListTaskLinks(ctx context.Context, taskID, businessID int64) ([]domain.TaskLink, error) {
	rows, err := r.queries.ListTaskLinks(ctx, sqlc.ListTaskLinksParams{TaskID: taskID, BusinessID: businessID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TaskLink, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTaskLink(row))
	}
	return out, nil
}

func (r *Repository) DeleteTaskLink(ctx context.Context, id, businessID int64) error {
	return r.queries.DeleteTaskLink(ctx, sqlc.DeleteTaskLinkParams{ID: id, BusinessID: businessID})
}
