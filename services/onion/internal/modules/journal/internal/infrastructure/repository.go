package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.JournalEntry, error) {
	query := `
		SELECT id, owner_id, date, description, created_at
		FROM finance_journal_entries
		WHERE owner_id = ?
		ORDER BY date DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list journal entries", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var entries []domain.JournalEntry
	for rows.Next() {
		var entry domain.JournalEntry
		if err := entry.FromRow(rows); err != nil {
			return nil, err
		}

		lines, err := r.GetLines(ctx, entry.ID)
		if err != nil {
			return nil, err
		}
		entry.Lines = lines

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *Repository) GetLines(ctx context.Context, entryID int64) ([]domain.JournalLine, error) {
	query := `
		SELECT id, entry_id, account_id, debit, credit
		FROM finance_journal_lines
		WHERE entry_id = ?
		ORDER BY id ASC`

	rows, err := r.pool.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []domain.JournalLine
	for rows.Next() {
		var line domain.JournalLine
		if err := line.FromRow(rows); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error) {
	query := `
		INSERT INTO finance_journal_entries (owner_id, date, description)
		VALUES (?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.Date,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create journal entry", logger.Error(err))
		return 0, err
	}

	entryID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, line := range input.Lines {
		_, err := r.pool.ExecContext(ctx,
			`INSERT INTO finance_journal_lines (entry_id, account_id, debit, credit) VALUES (?, ?, ?, ?)`,
			entryID, line.AccountID, line.Debit, line.Credit,
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create journal line", logger.Error(err))
			return 0, err
		}
	}

	return entryID, nil
}
