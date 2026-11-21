package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	"github.com/voxire/onion/internal/modules/pager/internal/infrastructure/persistence/sqlc"
)

const mysqlDuplicateEntry = 1062

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) Settings(ctx context.Context, businessID int64) (domain.Settings, error) {
	row, err := r.queries.GetSettings(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.Settings{Enabled: row.Enabled, LabelMode: row.LabelMode}, nil
}

func (r *Repository) SaveSettings(ctx context.Context, businessID int64, settings domain.Settings) error {
	return r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		BusinessID: businessID,
		Enabled:    settings.Enabled,
		LabelMode:  settings.LabelMode,
	})
}

// CreateTicket relies on uk_pager_tickets_open_match_code: a second open
// unlabelled ticket with the same code is refused by the database itself.
func (r *Repository) CreateTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	result, err := r.queries.CreateTicket(ctx, sqlc.CreateTicketParams{
		BusinessID:  ticket.BusinessID,
		MatchCode:   sql.NullString{String: ticket.Code, Valid: true},
		PublicToken: ticket.PublicToken,
		CreatedAt:   ticket.CreatedAt,
	})
	if isDuplicate(err) {
		return domain.Ticket{}, domain.ErrCodeTaken
	}
	if err != nil {
		return domain.Ticket{}, err
	}
	ticket.ID, err = result.LastInsertId()
	if err != nil {
		return domain.Ticket{}, err
	}
	ticket.Status = domain.StatusWaiting
	return ticket, nil
}

func (r *Repository) OpenMatchCodes(ctx context.Context, businessID int64, unlabelledSince time.Time) ([]string, error) {
	rows, err := r.queries.ListOpenMatchCodes(ctx, sqlc.ListOpenMatchCodesParams{
		BusinessID:      businessID,
		UnlabelledSince: unlabelledSince,
	})
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.String)
	}
	return codes, nil
}

func (r *Repository) ExpireStaleUnlabelled(ctx context.Context, businessID int64, unlabelledSince time.Time) error {
	return r.queries.ExpireStaleUnlabelledTickets(ctx, sqlc.ExpireStaleUnlabelledTicketsParams{
		BusinessID:      businessID,
		UnlabelledSince: unlabelledSince,
	})
}

func (r *Repository) ExpireStaleTicket(ctx context.Context, businessID int64, number string, openSince time.Time) error {
	return r.queries.ExpireStaleTicket(ctx, sqlc.ExpireStaleTicketParams{
		BusinessID: businessID,
		OpenNumber: sql.NullString{String: number, Valid: true},
		OpenSince:  openSince,
	})
}

func (r *Repository) ListOpenTickets(ctx context.Context, businessID int64, openSince, unlabelledSince time.Time) ([]domain.Ticket, error) {
	rows, err := r.queries.ListOpenTickets(ctx, sqlc.ListOpenTicketsParams{
		BusinessID:      businessID,
		OpenSince:       openSince,
		UnlabelledSince: unlabelledSince,
	})
	if err != nil {
		return nil, err
	}
	tickets := make([]domain.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, ticketFromRow(sqlc.GetTicketRow(row)))
	}
	return tickets, nil
}

func (r *Repository) Ticket(ctx context.Context, businessID, id int64) (domain.Ticket, bool, error) {
	return found(r.queries.GetTicket(ctx, sqlc.GetTicketParams{ID: id, BusinessID: businessID}))
}

func (r *Repository) TicketByToken(ctx context.Context, token string) (domain.Ticket, bool, error) {
	row, err := r.queries.GetTicketByToken(ctx, token)
	return found(sqlc.GetTicketRow(row), err)
}

func (r *Repository) CountWaitingAhead(ctx context.Context, businessID, id int64, openSince time.Time) (int64, error) {
	return r.queries.CountWaitingAhead(ctx, sqlc.CountWaitingAheadParams{
		BusinessID: businessID,
		OpenSince:  openSince,
		ID:         id,
	})
}

// Label relies on uk_pager_tickets_open_number for a number. Only a number
// fills the number column, so names repeat freely.
func (r *Repository) Label(ctx context.Context, businessID, id int64, label, kind string, openSince, unlabelledSince time.Time) error {
	err := r.queries.LabelTicket(ctx, sqlc.LabelTicketParams{
		Label:           sql.NullString{String: label, Valid: true},
		LabelKind:       sql.NullString{String: kind, Valid: true},
		Number:          sql.NullString{String: label, Valid: kind == domain.LabelNumber},
		ID:              id,
		BusinessID:      businessID,
		OpenSince:       openSince,
		UnlabelledSince: unlabelledSince,
	})
	if isDuplicate(err) {
		return domain.ErrLabelTaken
	}
	return err
}

func (r *Repository) MarkReady(ctx context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.queries.MarkTicketReady(ctx, sqlc.MarkTicketReadyParams{
		ReadyAt:    sql.NullTime{Time: at, Valid: true},
		ID:         id,
		BusinessID: businessID,
		OpenSince:  openSince,
	})
}

func (r *Repository) Remind(ctx context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.queries.RemindTicket(ctx, sqlc.RemindTicketParams{
		RemindedAt: sql.NullTime{Time: at, Valid: true},
		ID:         id,
		BusinessID: businessID,
		OpenSince:  openSince,
	})
}

func (r *Repository) UndoReady(ctx context.Context, businessID, id int64, openSince time.Time) error {
	return r.queries.UndoTicketReady(ctx, sqlc.UndoTicketReadyParams{ID: id, BusinessID: businessID, OpenSince: openSince})
}

func (r *Repository) Collect(ctx context.Context, businessID, id int64, openSince, unlabelledSince time.Time) error {
	return r.queries.CollectTicket(ctx, sqlc.CollectTicketParams{
		ID:              id,
		BusinessID:      businessID,
		OpenSince:       openSince,
		UnlabelledSince: unlabelledSince,
	})
}

func (r *Repository) Acknowledge(ctx context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.queries.AcknowledgeTicket(ctx, sqlc.AcknowledgeTicketParams{
		AcknowledgedAt: sql.NullTime{Time: at, Valid: true},
		ID:             id,
		BusinessID:     businessID,
		OpenSince:      openSince,
	})
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry
}

// found turns a missing row into "not found" rather than an error.
func found(row sqlc.GetTicketRow, err error) (domain.Ticket, bool, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Ticket{}, false, nil
	}
	if err != nil {
		return domain.Ticket{}, false, err
	}
	return ticketFromRow(row), true, nil
}

// ticketFromRow maps a ticket row. Every ticket query selects the same
// columns, so their row types convert to sqlc.GetTicketRow.
func ticketFromRow(row sqlc.GetTicketRow) domain.Ticket {
	return domain.Ticket{
		ID:             row.ID,
		BusinessID:     row.BusinessID,
		Label:          row.Label.String,
		LabelKind:      row.LabelKind.String,
		Code:           row.MatchCode.String,
		Status:         row.Status,
		PublicToken:    row.PublicToken,
		CreatedAt:      row.CreatedAt,
		ReadyAt:        row.ReadyAt.Time,
		RemindedAt:     row.RemindedAt.Time,
		AcknowledgedAt: row.AcknowledgedAt.Time,
	}
}
