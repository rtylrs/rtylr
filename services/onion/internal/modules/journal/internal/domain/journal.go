package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrJournalEntryNotFound      = errors.New("journal entry not found")
	ErrJournalEntryAlreadyPosted = errors.New("journal entry already posted")
)

type JournalEntry struct {
	ID          int64         `json:"id"`
	OwnerID     int64         `json:"owner_id"`
	Date        time.Time     `json:"date"`
	Description string        `json:"description"`
	Status      string        `json:"status"`
	PostedAt    time.Time     `json:"posted_at"`
	Lines       []JournalLine `json:"lines"`
	CreatedAt   time.Time     `json:"created_at"`
}

func (j *JournalEntry) FromRow(row *sql.Rows) error {
	var (
		description sql.NullString
		status      sql.NullString
		postedAt    sql.NullTime
	)

	err := row.Scan(&j.ID, &j.OwnerID, &j.Date, &description, &status, &postedAt, &j.CreatedAt)
	if err != nil {
		return err
	}

	if description.Valid {
		j.Description = description.String
	}
	if status.Valid {
		j.Status = status.String
	}
	if postedAt.Valid {
		j.PostedAt = postedAt.Time
	}

	return nil
}

type JournalLine struct {
	ID        int64   `json:"id"`
	EntryID   int64   `json:"entry_id"`
	AccountID int64   `json:"account_id"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

func (l *JournalLine) FromRow(row *sql.Rows) error {
	return row.Scan(&l.ID, &l.EntryID, &l.AccountID, &l.Debit, &l.Credit)
}

type JournalEntryInput struct {
	Date        string             `json:"date"`
	Description string             `json:"description"`
	Lines       []JournalLineInput `json:"lines"`
}

type JournalLineInput struct {
	AccountID int64   `json:"account_id"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}
