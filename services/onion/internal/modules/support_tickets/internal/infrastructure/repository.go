package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
	"github.com/voxire/onion/internal/modules/support_tickets/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
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
			Offset:     int32(offset),
		})
	} else {
		rows, err = r.queries.ListTickets(ctx, sqlc.ListTicketsParams{
			BusinessID: businessID,
			Limit:      int32(pageSize),
			Offset:     int32(offset),
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
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get support ticket", logger.Error(err))
		return nil, err
	}
	t := ticketFromRow(row)
	return &t, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TicketInput) (int64, error) {
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
	result, err := r.queries.CreateTicket(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create support ticket", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TicketUpdateInput) error {
	params := sqlc.UpdateTicketParams{
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

func (r *Repository) ListMessages(ctx context.Context, ticketID int64) ([]domain.Message, error) {
	rows, err := r.queries.ListMessages(ctx, ticketID)
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
	result, err := r.queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		TicketID:   input.TicketID,
		SenderType: sqlc.SupportMessagesSenderType(input.SenderType),
		SenderName: input.SenderName,
		Body:       input.Body,
		IsInternal: sql.NullBool{Bool: input.IsInternal, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create support message", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
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
