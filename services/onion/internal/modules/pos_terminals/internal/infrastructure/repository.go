package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_terminals/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Terminal, error) {
	rows, err := r.queries.ListTerminalsByBusiness(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list terminals", logger.Error(err))
		return nil, err
	}

	terminals := make([]domain.Terminal, 0, len(rows))
	for _, row := range rows {
		t := domain.Terminal{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			Name:       row.Name,
			IsActive:   row.IsActive.Bool,
			CreatedAt:  row.CreatedAt.Time,
		}
		if row.DeviceID.Valid {
			t.DeviceID = &row.DeviceID.String
		}
		if row.LastSeenAt.Valid {
			ts := row.LastSeenAt.Time
			t.LastSeenAt = &ts
		}
		terminals = append(terminals, t)
	}
	return terminals, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TerminalInput) (int64, error) {
	result, err := r.queries.CreateTerminal(ctx, sqlc.CreateTerminalParams{
		BusinessID: input.BusinessID,
		BranchID:   input.BranchID,
		Name:       input.Name,
		DeviceID:   nullStringPtr(input.DeviceID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create terminal", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TerminalUpdateInput) error {
	err := r.queries.UpdateTerminal(ctx, sqlc.UpdateTerminalParams{
		BranchID:   input.BranchID,
		Name:       input.Name,
		DeviceID:   nullStringPtr(input.DeviceID),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:         input.ID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update terminal", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Heartbeat(ctx context.Context, terminalID, businessID int64) error {
	result, err := r.queries.UpdateTerminalHeartbeat(ctx, sqlc.UpdateTerminalHeartbeatParams{
		ID:         terminalID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update terminal heartbeat", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("terminal not found")
	}
	return nil
}

func nullStringPtr(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}
