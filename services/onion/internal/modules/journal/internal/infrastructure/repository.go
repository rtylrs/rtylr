package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.JournalEntry, error) {
	rows, err := r.queries.ListJournalEntries(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list journal entries", logger.Error(err))
		return nil, err
	}

	var entries []domain.JournalEntry
	for _, row := range rows {
		entry := domain.JournalEntry{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Date:        row.Date,
			Description: nullStr(row.Description),
			CreatedAt:   nullTime(row.CreatedAt),
		}

		lines, err := r.GetLines(ctx, entry.ID)
		if err != nil {
			return nil, err
		}
		entry.Lines = lines

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *Repository) GetLines(ctx context.Context, entryID int64) ([]domain.JournalLine, error) {
	rows, err := r.queries.GetJournalLines(ctx, entryID)
	if err != nil {
		return nil, err
	}

	lines := make([]domain.JournalLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, domain.JournalLine{
			ID:        row.ID,
			EntryID:   row.EntryID,
			AccountID: row.AccountID,
			Debit:     parseFloatNull(row.Debit),
			Credit:    parseFloatNull(row.Credit),
		})
	}

	return lines, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.JournalEntryInput) (int64, error) {
	entryDate, _ := time.Parse("2006-01-02", input.Date)

	res, err := r.queries.CreateJournalEntry(ctx, sqlc.CreateJournalEntryParams{
		OwnerID:     ownerID,
		Date:        entryDate,
		Description: nullableString(input.Description),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create journal entry", logger.Error(err))
		return 0, err
	}

	entryID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, line := range input.Lines {
		err := r.queries.CreateJournalLine(ctx, sqlc.CreateJournalLineParams{
			EntryID:   entryID,
			AccountID: line.AccountID,
			Debit:     nullableString(fmtFloat(line.Debit)),
			Credit:    nullableString(fmtFloat(line.Credit)),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create journal line", logger.Error(err))
			return 0, err
		}
	}

	return entryID, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	v, _ := strconv.ParseFloat(s.String, 64)
	return v
}
