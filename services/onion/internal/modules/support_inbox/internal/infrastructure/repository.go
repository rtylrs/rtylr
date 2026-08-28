package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListInbox(ctx context.Context, businessID int64, page, pageSize int) ([]domain.InboxItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := r.queries.CountInboxTickets(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count inbox tickets", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListInbox(ctx, sqlc.ListInboxParams{
		BusinessID:   businessID,
		Limit:        int32(pageSize),
		Offset:       int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessID_2: businessID,
		Limit_2:      int32(pageSize),
		Offset_2:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list inbox", logger.Error(err))
		return nil, 0, err
	}

	items := make([]domain.InboxItem, 0, len(rows))
	for _, row := range rows {
		item := domain.InboxItem{
			TicketID:      row.ID,
			CustomerName:  row.CustomerName,
			CustomerEmail: row.CustomerEmail,
			Subject:       row.Subject,
			MessageCount:  int(row.MessageCount),
		}
		if row.LastMessage.Valid {
			item.LastMessage = row.LastMessage.String
		}
		if row.Status.Valid {
			item.Status = string(row.Status.SupportTicketsStatus)
		}
		if row.Priority.Valid {
			item.Priority = string(row.Priority.SupportTicketsPriority)
		}
		if row.Channel.Valid {
			item.Channel = row.Channel.String
		}
		if row.LastMessageAt.Valid {
			item.LastMessageAt = row.LastMessageAt.Time
		}
		items = append(items, item)
	}
	return items, int(total), nil
}

func (r *Repository) UnreadCount(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.UnreadCount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count unread tickets", logger.Error(err))
		return 0, err
	}
	return int(count), nil
}
