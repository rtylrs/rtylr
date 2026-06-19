package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetWidgets(ctx context.Context, businessID int64) ([]*domain.DashboardWidget, error) {
	query, args, err := sq.
		Select(
			"id", "owner_id", "business_id", "role_id", "widget_type",
			"position_x", "position_y", "width", "height",
			"config", "created_at", "updated_at",
		).
		From("hr_dashboard_widgets").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("position_y ASC", "position_x ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := make([]*domain.DashboardWidget, 0)
	for rows.Next() {
		var w domain.DashboardWidget
		var configRaw sql.NullString
		err := rows.Scan(
			&w.ID, &w.OwnerID, &w.BusinessID, &w.RoleID, &w.WidgetType,
			&w.PositionX, &w.PositionY, &w.Width, &w.Height,
			&configRaw, &w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if configRaw.Valid {
			w.Config = []byte(configRaw.String)
		} else {
			w.Config = []byte("{}")
		}
		widgets = append(widgets, &w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return widgets, nil
}

func (r *Repository) UpsertWidgets(ctx context.Context, businessID int64, widgets []*domain.DashboardWidget) error {
	tx, err := r.db.BeginTx(ctx, nil)
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
		// Batch every widget into a single multi-row INSERT instead of issuing
		// one round-trip per widget.
		insert := sq.
			Insert("hr_dashboard_widgets").
			Columns("owner_id", "business_id", "role_id", "widget_type", "position_x", "position_y", "width", "height", "config")
		for _, w := range widgets {
			configJSON, err := json.Marshal(json.RawMessage(w.Config))
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			insert = insert.Values(businessID, businessID, w.RoleID, w.WidgetType, w.PositionX, w.PositionY, w.Width, w.Height, configJSON)
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
