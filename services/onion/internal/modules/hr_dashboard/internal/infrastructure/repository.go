package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

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

func (r *Repository) GetWidgets(ctx context.Context, ownerID int64) ([]*domain.DashboardWidget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_id, role_id, widget_type, position_x, position_y, width, height, config, created_at, updated_at
		 FROM hr_dashboard_widgets
		 WHERE owner_id = ?
		 ORDER BY position_y ASC, position_x ASC`,
		ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := make([]*domain.DashboardWidget, 0)
	for rows.Next() {
		var w domain.DashboardWidget
		var configRaw sql.NullString
		err := rows.Scan(
			&w.ID, &w.OwnerID, &w.RoleID, &w.WidgetType,
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

func (r *Repository) UpsertWidgets(ctx context.Context, ownerID int64, widgets []*domain.DashboardWidget) error {
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

	_, err = tx.ExecContext(ctx, `DELETE FROM hr_dashboard_widgets WHERE owner_id = ?`, ownerID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if len(widgets) > 0 {
		// Batch every widget into a single multi-row INSERT instead of issuing
		// one round-trip per widget.
		placeholders := make([]string, 0, len(widgets))
		args := make([]interface{}, 0, len(widgets)*8)
		for _, w := range widgets {
			configJSON, err := json.Marshal(json.RawMessage(w.Config))
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args, ownerID, w.RoleID, w.WidgetType, w.PositionX, w.PositionY, w.Width, w.Height, configJSON)
		}

		query := `INSERT INTO hr_dashboard_widgets (owner_id, role_id, widget_type, position_x, position_y, width, height, config) VALUES ` +
			strings.Join(placeholders, ", ")
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
