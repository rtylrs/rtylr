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
	return domain.Settings{
		Enabled:         row.Enabled,
		LabelMode:       row.LabelMode,
		ReviewURL:       row.ReviewUrl.String,
		PrepMinutes:     int(row.PrepMinutes),
		CustomerNumbers: row.CustomerNumbers,
		Ringtone:        row.Ringtone,
	}, nil
}

func (r *Repository) SaveSettings(ctx context.Context, businessID int64, settings domain.Settings) error {
	return r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		BusinessID:      businessID,
		Enabled:         settings.Enabled,
		LabelMode:       settings.LabelMode,
		ReviewUrl:       sql.NullString{String: settings.ReviewURL, Valid: settings.ReviewURL != ""},
		PrepMinutes:     int16(settings.PrepMinutes),
		CustomerNumbers: settings.CustomerNumbers,
		Ringtone:        settings.Ringtone,
	})
}

// CreateNumberTicket opens a ticket labelled with the number a customer typed.
// uk_pager_tickets_open_number refuses one that already has an open ticket.
func (r *Repository) CreateNumberTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	result, err := r.queries.CreateNumberTicket(ctx, sqlc.CreateNumberTicketParams{
		BusinessID:      ticket.BusinessID,
		Number:          sql.NullString{String: ticket.Label, Valid: true},
		Label:           sql.NullString{String: ticket.Label, Valid: true},
		PublicToken:     ticket.PublicToken,
		CreatedAt:       ticket.CreatedAt,
		ExpectedReadyAt: nullTime(ticket.ExpectedReadyAt),
	})
	if isDuplicate(err) {
		return domain.Ticket{}, domain.ErrLabelTaken
	}
	if err != nil {
		return domain.Ticket{}, err
	}
	ticket.ID, err = result.LastInsertId()
	if err != nil {
		return domain.Ticket{}, err
	}
	ticket.LabelKind = domain.LabelNumber
	ticket.Status = domain.StatusWaiting
	return ticket, nil
}

func (r *Repository) CountOpenTickets(ctx context.Context, businessID int64, openSince, unlabelledSince time.Time) (int64, error) {
	return r.queries.CountOpenTickets(ctx, sqlc.CountOpenTicketsParams{
		BusinessID:      businessID,
		OpenSince:       openSince,
		UnlabelledSince: unlabelledSince,
	})
}

func (r *Repository) TicketTimes(ctx context.Context, businessID int64, since time.Time) ([]domain.TicketTimes, error) {
	rows, err := r.queries.ListLabelledTicketTimes(ctx, sqlc.ListLabelledTicketTimesParams{
		BusinessID: businessID,
		Since:      since,
	})
	if err != nil {
		return nil, err
	}
	times := make([]domain.TicketTimes, 0, len(rows))
	for _, row := range rows {
		times = append(times, domain.TicketTimes{CreatedAt: row.CreatedAt, ReadyAt: row.ReadyAt.Time})
	}
	return times, nil
}

// CreateTicket relies on uk_pager_tickets_open_match_number: a second open
// unlabelled ticket with the same number is refused by the database itself.
func (r *Repository) CreateTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	result, err := r.queries.CreateTicket(ctx, sqlc.CreateTicketParams{
		BusinessID:  ticket.BusinessID,
		MatchNumber: sql.NullString{String: ticket.Code, Valid: true},
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
	return rows, nil
}

func (r *Repository) ExpireStaleUnlabelled(ctx context.Context, businessID int64, unlabelledSince time.Time) error {
	ids, err := r.queries.ListStaleUnlabelledTicketIDs(ctx, sqlc.ListStaleUnlabelledTicketIDsParams{
		BusinessID:      businessID,
		UnlabelledSince: unlabelledSince,
	})
	if err != nil {
		return err
	}
	for _, id := range ids {
		err := r.queries.ExpireStaleUnlabelledTicket(ctx, sqlc.ExpireStaleUnlabelledTicketParams{
			ID:              id,
			BusinessID:      businessID,
			UnlabelledSince: unlabelledSince,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) BusinessIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListPagerBusinessIDs(ctx)
}

// ExpireStaleLabelled goes through a plain read first, as ExpireStaleUnlabelled
// does, so it never locks a venue's range of open tickets.
func (r *Repository) ExpireStaleLabelled(ctx context.Context, businessID int64, openSince time.Time) error {
	ids, err := r.queries.ListStaleLabelledTicketIDs(ctx, sqlc.ListStaleLabelledTicketIDsParams{
		BusinessID: businessID,
		OpenSince:  openSince,
	})
	if err != nil {
		return err
	}
	for _, id := range ids {
		err := r.queries.ExpireStaleLabelledTicket(ctx, sqlc.ExpireStaleLabelledTicketParams{
			ID:         id,
			BusinessID: businessID,
			OpenSince:  openSince,
		})
		if err != nil {
			return err
		}
	}
	return nil
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
func (r *Repository) Label(ctx context.Context, businessID, id int64, label, kind string, expectedReadyAt, openSince, unlabelledSince time.Time) error {
	err := r.queries.LabelTicket(ctx, sqlc.LabelTicketParams{
		Label:           sql.NullString{String: label, Valid: true},
		LabelKind:       sql.NullString{String: kind, Valid: true},
		Number:          sql.NullString{String: label, Valid: kind == domain.LabelNumber},
		ExpectedReadyAt: nullTime(expectedReadyAt),
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

// matchOf is the ticket's match number, or on a ticket opened before the
// numbers, its word or two character code.
func matchOf(row sqlc.GetTicketRow) string {
	switch {
	case row.MatchNumber.Valid:
		return row.MatchNumber.String
	case row.MatchWord.Valid:
		return row.MatchWord.String
	}
	return row.MatchCode.String
}

// ticketFromRow maps a ticket row. Every ticket query selects the same
// columns, so their row types convert to sqlc.GetTicketRow.
func ticketFromRow(row sqlc.GetTicketRow) domain.Ticket {
	return domain.Ticket{
		ID:              row.ID,
		BusinessID:      row.BusinessID,
		Label:           row.Label.String,
		LabelKind:       row.LabelKind.String,
		Code:            matchOf(row),
		Status:          row.Status,
		PublicToken:     row.PublicToken,
		CreatedAt:       row.CreatedAt,
		ReadyAt:         row.ReadyAt.Time,
		RemindedAt:      row.RemindedAt.Time,
		AcknowledgedAt:  row.AcknowledgedAt.Time,
		ExpectedReadyAt: row.ExpectedReadyAt.Time,
	}
}

// nullTime stores a zero time as NULL.
func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}
