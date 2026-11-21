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
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.Settings{Enabled: row.Enabled, Strict: row.Strict}, nil
}

func (r *Repository) SaveSettings(ctx context.Context, businessID int64, settings domain.Settings) error {
	return r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		BusinessID: businessID,
		Enabled:    settings.Enabled,
		Strict:     settings.Strict,
	})
}

// CreateTicket relies on uk_pager_tickets_open_number: a second open ticket
// for the same business and number is refused by the database itself.
func (r *Repository) CreateTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	result, err := r.queries.CreateTicket(ctx, sqlc.CreateTicketParams{
		BusinessID:  ticket.BusinessID,
		Number:      ticket.Number,
		Source:      ticket.Source,
		PublicToken: ticket.PublicToken,
		PhoneLinked: ticket.PhoneLinked,
		CreatedAt:   ticket.CreatedAt,
	})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
		return domain.Ticket{}, domain.ErrNumberTaken
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

func (r *Repository) OpenTicketByNumber(ctx context.Context, businessID int64, number string, openSince time.Time) (domain.Ticket, bool, error) {
	row, err := r.queries.GetOpenTicketByNumber(ctx, sqlc.GetOpenTicketByNumberParams{
		BusinessID: businessID,
		OpenNumber: sql.NullString{String: number, Valid: true},
		OpenSince:  openSince,
	})
	return found(sqlc.GetTicketRow(row), err)
}

func (r *Repository) ExpireStaleTicket(ctx context.Context, businessID int64, number string, openSince time.Time) error {
	return r.queries.ExpireStaleTicket(ctx, sqlc.ExpireStaleTicketParams{
		BusinessID: businessID,
		OpenNumber: sql.NullString{String: number, Valid: true},
		OpenSince:  openSince,
	})
}

func (r *Repository) CountOpenCustomerTickets(ctx context.Context, businessID int64, openSince time.Time) (int64, error) {
	return r.queries.CountOpenCustomerTickets(ctx, sqlc.CountOpenCustomerTicketsParams{
		BusinessID: businessID,
		OpenSince:  openSince,
	})
}

func (r *Repository) ListOpenTickets(ctx context.Context, businessID int64, openSince time.Time) ([]domain.Ticket, error) {
	rows, err := r.queries.ListOpenTickets(ctx, sqlc.ListOpenTicketsParams{
		BusinessID: businessID,
		OpenSince:  openSince,
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

func (r *Repository) LinkPhone(ctx context.Context, businessID, id int64) error {
	return r.queries.LinkPhone(ctx, sqlc.LinkPhoneParams{ID: id, BusinessID: businessID})
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

func (r *Repository) Collect(ctx context.Context, businessID, id int64, openSince time.Time) error {
	return r.queries.CollectTicket(ctx, sqlc.CollectTicketParams{ID: id, BusinessID: businessID, OpenSince: openSince})
}

func (r *Repository) Acknowledge(ctx context.Context, businessID, id int64, at, openSince time.Time) error {
	return r.queries.AcknowledgeTicket(ctx, sqlc.AcknowledgeTicketParams{
		AcknowledgedAt: sql.NullTime{Time: at, Valid: true},
		ID:             id,
		BusinessID:     businessID,
		OpenSince:      openSince,
	})
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
		Number:         row.Number,
		Status:         row.Status,
		Source:         row.Source,
		PublicToken:    row.PublicToken,
		PhoneLinked:    row.PhoneLinked,
		CreatedAt:      row.CreatedAt,
		ReadyAt:        row.ReadyAt.Time,
		RemindedAt:     row.RemindedAt.Time,
		AcknowledgedAt: row.AcknowledgedAt.Time,
	}
}
