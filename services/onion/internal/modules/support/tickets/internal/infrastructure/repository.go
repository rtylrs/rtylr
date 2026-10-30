package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	supportcontract "github.com/voxire/onion/internal/modules/support/contract"
	"github.com/voxire/onion/internal/modules/support/internal/outbox"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64, status string, page, pageSize int) ([]domain.Ticket, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	var err error

	if status != "" {
		total, err = r.queries.CountTicketsByStatus(ctx, sqlc.CountTicketsByStatusParams{
			BusinessID: businessID,
			Status:     sqlc.NullSupportTicketsStatus{SupportTicketsStatus: sqlc.SupportTicketsStatus(status), Valid: true},
		})
	} else {
		total, err = r.queries.CountTickets(ctx, businessID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count support tickets", logger.Error(err))
		return nil, 0, err
	}

	var rows []sqlc.SupportTicket
	if status != "" {
		rows, err = r.queries.ListTicketsByStatus(ctx, sqlc.ListTicketsByStatusParams{
			BusinessID: businessID,
			Status:     sqlc.NullSupportTicketsStatus{SupportTicketsStatus: sqlc.SupportTicketsStatus(status), Valid: true},
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	} else {
		rows, err = r.queries.ListTickets(ctx, sqlc.ListTicketsParams{
			BusinessID: businessID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list support tickets", logger.Error(err))
		return nil, 0, err
	}

	items := make([]domain.Ticket, 0, len(rows))
	for _, row := range rows {
		items = append(items, ticketFromRow(row))
	}
	return items, int(total), nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (*domain.Ticket, error) {
	row, err := r.queries.GetTicketByID(ctx, sqlc.GetTicketByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTicketNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get support ticket", logger.Error(err))
		return nil, err
	}
	t := ticketFromRow(row)
	return &t, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TicketInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin support ticket transaction", logger.Error(err))
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = upsertSupportCustomer(ctx, r.queries.WithTx(tx), input.BusinessID, input.CustomerName, input.CustomerEmail); err != nil {
		logger.Ctx(ctx).Error("Failed to sync support customer summary", logger.Error(err))
		return 0, err
	}

	params := sqlc.CreateTicketParams{
		BusinessID:    input.BusinessID,
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		Subject:       input.Subject,
	}
	if input.Status != "" {
		params.Status = sqlc.NullSupportTicketsStatus{SupportTicketsStatus: sqlc.SupportTicketsStatus(input.Status), Valid: true}
	}
	if input.Priority != "" {
		params.Priority = sqlc.NullSupportTicketsPriority{SupportTicketsPriority: sqlc.SupportTicketsPriority(input.Priority), Valid: true}
	}
	if input.AssignedTo != nil {
		params.AssignedTo = sql.NullInt64{Int64: *input.AssignedTo, Valid: true}
	}
	if input.Channel != "" {
		params.Channel = sql.NullString{String: input.Channel, Valid: true}
	}
	result, err := r.queries.WithTx(tx).CreateTicket(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create support ticket", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  input.BusinessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "support_ticket",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("support_ticket", id, contact),
	}); err != nil {
		return 0, err
	}
	if err = outbox.Publish(ctx, tx, eventbus.Event{BusinessID: input.BusinessID, Type: supportcontract.TicketCreated, SubjectType: "support_ticket", SubjectID: id}); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit support ticket transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func upsertSupportCustomer(ctx context.Context, qtx *sqlc.Queries, businessID int64, name, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	customerID, err := qtx.FindSupportCustomerByEmail(ctx, sqlc.FindSupportCustomerByEmailParams{
		BusinessID: businessID,
		Email:      sql.NullString{String: email, Valid: email != ""},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return qtx.InsertSupportCustomer(ctx, sqlc.InsertSupportCustomerParams{
			BusinessID: businessID,
			Name:       name,
			Email:      sql.NullString{String: email, Valid: email != ""},
		})
	}
	if err != nil {
		return err
	}

	return qtx.IncrementSupportCustomer(ctx, sqlc.IncrementSupportCustomerParams{
		Column1:    name,
		Name:       name,
		Column3:    email,
		Email:      sql.NullString{String: email, Valid: email != ""},
		BusinessID: businessID,
		ID:         customerID,
	})
}

func (r *Repository) Update(ctx context.Context, input domain.TicketUpdateInput) error {
	// The statement treats an omitted field (empty string / NULL) as "keep the
	// stored value", so the two subject parameters carry the same value.
	params := sqlc.UpdateTicketParams{
		Column1:    input.Subject,
		Subject:    input.Subject,
		BusinessID: input.BusinessID,
		ID:         input.ID,
	}
	if input.Status != "" {
		params.Status = sqlc.NullSupportTicketsStatus{SupportTicketsStatus: sqlc.SupportTicketsStatus(input.Status), Valid: true}
	}
	if input.Priority != "" {
		params.Priority = sqlc.NullSupportTicketsPriority{SupportTicketsPriority: sqlc.SupportTicketsPriority(input.Priority), Valid: true}
	}
	if input.AssignedTo != nil {
		params.AssignedTo = sql.NullInt64{Int64: *input.AssignedTo, Valid: true}
	}
	if input.Channel != "" {
		params.Channel = sql.NullString{String: input.Channel, Valid: true}
	}
	err := r.queries.UpdateTicket(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update support ticket", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ListMessages(ctx context.Context, businessID, ticketID int64) ([]domain.Message, error) {
	rows, err := r.queries.ListMessages(ctx, sqlc.ListMessagesParams{TicketID: ticketID, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list support messages", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Message{
			ID:         row.ID,
			TicketID:   row.TicketID,
			SenderType: string(row.SenderType),
			SenderName: row.SenderName,
			Body:       row.Body,
			IsInternal: row.IsInternal.Bool,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) CreateMessage(ctx context.Context, input domain.MessageInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)
	businessID, err := q.GetTicketBusiness(ctx, input.TicketID)
	if err != nil {
		return 0, err
	}
	result, err := q.CreateMessage(ctx, sqlc.CreateMessageParams{
		TicketID: input.TicketID, SenderType: sqlc.SupportMessagesSenderType(input.SenderType),
		SenderName: input.SenderName, Body: input.Body,
		IsInternal: sql.NullBool{Bool: input.IsInternal, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if input.SenderType == "customer" && !input.IsInternal {
		if err := outbox.Publish(ctx, tx, eventbus.Event{BusinessID: businessID, Type: supportcontract.CustomerMessageCreated, SubjectType: "support_message", SubjectID: id, Payload: map[string]any{"ticket_id": input.TicketID}}); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func ticketFromRow(row sqlc.SupportTicket) domain.Ticket {
	t := domain.Ticket{
		ID:            row.ID,
		BusinessID:    row.BusinessID,
		CustomerName:  row.CustomerName,
		CustomerEmail: row.CustomerEmail,
		Subject:       row.Subject,
	}
	if row.Status.Valid {
		t.Status = string(row.Status.SupportTicketsStatus)
	}
	if row.Priority.Valid {
		t.Priority = string(row.Priority.SupportTicketsPriority)
	}
	if row.AssignedTo.Valid {
		v := row.AssignedTo.Int64
		t.AssignedTo = &v
	}
	if row.Channel.Valid {
		t.Channel = row.Channel.String
	}
	if row.CreatedAt.Valid {
		t.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		t.UpdatedAt = row.UpdatedAt.Time
	}
	return t
}
