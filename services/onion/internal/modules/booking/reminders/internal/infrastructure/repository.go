package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking/reminders/internal/domain"
	"github.com/voxire/onion/internal/modules/booking/reminders/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Reminder, error) {
	rows, err := r.queries.ListReminders(ctx, businessID)
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

// AppointmentExists reports whether the given appointment exists and belongs to
// businessID. It guards reminder creation against referencing another tenant's
// appointment.
//
// cross-schema read from booking_appointments — verify the appointment belongs to the business before creating a reminder for it
func (r *Repository) AppointmentExists(ctx context.Context, businessID, appointmentID int64) (bool, error) {
	count, err := r.queries.CountAppointmentForOwner(ctx, sqlc.CountAppointmentForOwnerParams{
		ID:         appointmentID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify appointment ownership", logger.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.ReminderInput) (int64, error) {
	scheduledAt, err := time.Parse("2006-01-02T15:04:05Z07:00", input.ScheduledAt)
	if err != nil {
		// The handler validates scheduled_at is non-empty, so a parse failure
		// here indicates corrupt input rather than an optional empty value.
		logger.Ctx(ctx).Error("Failed to parse scheduled_at",
			logger.String("scheduled_at", input.ScheduledAt),
			logger.Error(err),
		)
		return 0, err
	}
	result, err := r.queries.CreateReminder(ctx, sqlc.CreateReminderParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		AppointmentID: sql.NullInt64{Int64: input.AppointmentID, Valid: input.AppointmentID != 0},
		Type:          sqlc.NullBookingRemindersType{BookingRemindersType: sqlc.BookingRemindersType(input.Type), Valid: input.Type != ""},
		ScheduledAt:   sql.NullTime{Time: scheduledAt, Valid: input.ScheduledAt != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create reminder", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read reminder insert id", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func fromRow(row sqlc.BookingReminder) domain.Reminder {
	rem := domain.Reminder{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID}
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
