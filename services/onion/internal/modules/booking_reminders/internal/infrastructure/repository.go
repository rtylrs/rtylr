package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_reminders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Reminder, error) {
	rows, err := r.queries.ListReminders(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reminders", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Reminder, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ReminderInput) (int64, error) {
	scheduledAt, err := time.Parse("2006-01-02T15:04:05Z07:00", input.ScheduledAt)
	if err != nil && input.ScheduledAt != "" {
		logger.Ctx(ctx).Error("Failed to parse scheduled_at", logger.Error(err))
		return 0, err
	}
	result, err := r.queries.CreateReminder(ctx, sqlc.CreateReminderParams{
		OwnerID:       ownerID,
		AppointmentID: sql.NullInt64{Int64: input.AppointmentID, Valid: input.AppointmentID != 0},
		Type:          sqlc.NullBookingRemindersType{BookingRemindersType: sqlc.BookingRemindersType(input.Type), Valid: input.Type != ""},
		ScheduledAt:   sql.NullTime{Time: scheduledAt, Valid: input.ScheduledAt != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create reminder", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func fromRow(row sqlc.BookingReminder) domain.Reminder {
	rem := domain.Reminder{ID: row.ID, OwnerID: row.OwnerID}
	if row.AppointmentID.Valid {
		rem.AppointmentID = row.AppointmentID.Int64
	}
	if row.Type.Valid {
		rem.Type = string(row.Type.BookingRemindersType)
	}
	if row.ScheduledAt.Valid {
		rem.ScheduledAt = row.ScheduledAt.Time
	}
	if row.SentAt.Valid {
		rem.SentAt = row.SentAt.Time
	}
	if row.CreatedAt.Valid {
		rem.CreatedAt = row.CreatedAt.Time
	}
	return rem
}
