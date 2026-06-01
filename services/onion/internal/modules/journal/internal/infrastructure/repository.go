package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	"github.com/voxire/onion/internal/modules/journal/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.JournalEntry, error) {
	rows, err := r.queries.ListJournalEntries(ctx, sqlc.ListJournalEntriesParams{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list journal entries", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.JournalEntry, 0, len(rows))
	entryIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, domain.JournalEntry{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Date:        row.Date,
			Description: nullStr(row.Description),
			Status:      row.Status,
			PostedAt:    nullTime(row.PostedAt),
			CreatedAt:   nullTime(row.CreatedAt),
		})
		entryIDs = append(entryIDs, row.ID)
	}

	// Batch-load all lines for the listed entries in a single query, then
	// attach them in memory. This replaces a per-entry GetLines call (N+1).
	linesByEntry, err := r.getLinesForEntries(ctx, entryIDs)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Lines = linesByEntry[entries[i].ID]
	}

	return entries, nil
}

// getLinesForEntries loads all journal lines for the given entry IDs in one
// query and groups them by entry ID. An entry with no lines maps to nil.
func (r *Repository) getLinesForEntries(ctx context.Context, entryIDs []int64) (map[int64][]domain.JournalLine, error) {
	out := make(map[int64][]domain.JournalLine, len(entryIDs))
	if len(entryIDs) == 0 {
		return out, nil
	}

	var sb strings.Builder
	sb.WriteString("SELECT id, entry_id, account_id, debit, credit FROM finance_journal_lines WHERE entry_id IN (")
	args := make([]any, len(entryIDs))
	for i, id := range entryIDs {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('?')
		args[i] = id
	}
	sb.WriteByte(')')

	dbRows, err := r.pool.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch-load journal lines", logger.Error(err))
		return nil, err
	}
	defer dbRows.Close()

	for dbRows.Next() {
		var (
			id, entryID, accountID int64
			debit, credit          sql.NullString
		)
		if scanErr := dbRows.Scan(&id, &entryID, &accountID, &debit, &credit); scanErr != nil {
			return nil, scanErr
		}
		out[entryID] = append(out[entryID], domain.JournalLine{
			ID:        id,
			EntryID:   entryID,
			AccountID: accountID,
			Debit:     parseFloatNull(debit),
			Credit:    parseFloatNull(credit),
		})
	}

	return out, dbRows.Err()
}

func (r *Repository) Get(ctx context.Context, ownerID, id int64) (domain.JournalEntry, error) {
	row, err := r.queries.GetJournalEntry(ctx, sqlc.GetJournalEntryParams{ID: id, OwnerID: ownerID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.JournalEntry{}, domain.ErrJournalEntryNotFound
		}
		logger.Ctx(ctx).Error("Failed to get journal entry", logger.Error(err))
		return domain.JournalEntry{}, err
	}

	return domain.JournalEntry{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		Date:        row.Date,
		Description: nullStr(row.Description),
		Status:      row.Status,
		PostedAt:    nullTime(row.PostedAt),
		CreatedAt:   nullTime(row.CreatedAt),
	}, nil
}

func (r *Repository) Post(ctx context.Context, ownerID, id int64) error {
	// Verify the entry exists and read its current status up-front so the
	// service can distinguish 404 from 409.
	entry, err := r.Get(ctx, ownerID, id)
	if err != nil {
		return err
	}
	if entry.Status == "posted" {
		return domain.ErrJournalEntryAlreadyPosted
	}

	res, err := r.queries.PostJournalEntry(ctx, sqlc.PostJournalEntryParams{ID: id, OwnerID: ownerID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to post journal entry",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
			logger.Any("entry_id", id),
		)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Raced with another caller — entry was posted between the read and
		// the update.
		return domain.ErrJournalEntryAlreadyPosted
	}

	return nil
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
	entryDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse journal entry date", logger.Error(err))
		return 0, err
	}

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

// CountByStatus returns how many journal entries the owner has for each
// status (e.g. "draft", "posted"). An empty map is returned when the owner
// has no entries.
func (r *Repository) CountByStatus(ctx context.Context, ownerID int64) (map[string]int64, error) {
	const q = `
		SELECT status, COUNT(*) AS cnt
		FROM finance_journal_entries
		WHERE owner_id = ?
		GROUP BY status`

	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count journal entries by status", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var s string
		var cnt int64
		if scanErr := rows.Scan(&s, &cnt); scanErr != nil {
			return nil, scanErr
		}
		counts[s] = cnt
	}

	return counts, rows.Err()
}

// SumLinesByAccountID returns the total debit and credit amounts posted to a
// given chart-of-accounts ID across all posted journal entries for an owner.
func (r *Repository) SumLinesByAccountID(ctx context.Context, ownerID, accountID int64) (debit, credit float64, err error) {
	const q = `
		SELECT
			COALESCE(SUM(CAST(l.debit AS REAL)), 0),
			COALESCE(SUM(CAST(l.credit AS REAL)), 0)
		FROM finance_journal_lines l
		JOIN finance_journal_entries e ON e.id = l.entry_id
		WHERE e.owner_id = ?
		  AND l.account_id = ?
		  AND e.status = 'posted'`

	err = r.pool.QueryRowContext(ctx, q, ownerID, accountID).Scan(&debit, &credit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum journal lines by account", logger.Error(err))
		return 0, 0, err
	}

	return debit, credit, nil
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
