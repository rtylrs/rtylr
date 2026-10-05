package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), pool: db}
}

func (r *Repository) GetWidgets(ctx context.Context, businessID int64) ([]*domain.DashboardWidget, error) {
	rows, err := r.queries.ListDashboardWidgets(ctx, businessID)
	if err != nil {
		return nil, err
	}
	widgets := make([]*domain.DashboardWidget, 0, len(rows))
	for _, row := range rows {
		w := &domain.DashboardWidget{
			ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID,
			RoleID: row.RoleID, WidgetType: row.WidgetType,
			PositionX: row.PositionX, PositionY: row.PositionY, Width: row.Width, Height: row.Height,
			CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		}
		if len(row.Config) > 0 {
			w.Config = row.Config
		} else {
			w.Config = []byte("{}")
		}
		widgets = append(widgets, w)
	}
	return widgets, nil
}

// dynamic-query: squirrel
func (r *Repository) UpsertWidgets(ctx context.Context, businessID int64, widgets []*domain.DashboardWidget) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	deleteQuery, deleteArgs, err := sq.
		Delete("hr_dashboard_widgets").
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if len(widgets) > 0 {
		insert := sq.
			Insert("hr_dashboard_widgets").
			Columns("owner_id", "business_id", "role_id", "widget_type", "position_x", "position_y", "width", "height", "config")
		for _, w := range widgets {
			configJSON, err := json.Marshal(json.RawMessage(w.Config))
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			insert = insert.Values(w.OwnerID, businessID, w.RoleID, w.WidgetType, w.PositionX, w.PositionY, w.Width, w.Height, configJSON)
		}

		query, args, err := insert.PlaceholderFormat(sq.Question).ToSql()
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
