package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.WaitlistEntry, error) {
	rows, err := r.queries.ListWaitlist(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list waitlist", logger.Error(err))
		return nil, err
	}
	items := make([]domain.WaitlistEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

// ServiceBelongsToOwner reports whether the given booking service exists and is
// owned by ownerID. It returns domain.ErrServiceNotFound when no matching row is
// found, so a caller cannot reference another tenant's service.
func (r *Repository) ServiceBelongsToOwner(ctx context.Context, ownerID, serviceID int64) error {
	// cross-schema read from booking_services — verify service ownership before linking a waitlist entry
	_, err := r.queries.GetBookingServiceForOwner(ctx, sqlc.GetBookingServiceForOwnerParams{
		ID:      serviceID,
		OwnerID: ownerID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrServiceNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify booking service ownership", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.WaitlistInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin waitlist transaction", logger.Error(err))
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, ownerID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
		Phone: input.CustomerPhone,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync waitlist customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateWaitlistEntry(ctx, sqlc.CreateWaitlistEntryParams{
		OwnerID:       ownerID,
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		PreferredDate: parseNullDate(input.PreferredDate),
		PreferredTime: parseNullTime(input.PreferredTime),
		Status:        sqlc.NullBookingWaitlistStatus{BookingWaitlistStatus: sqlc.BookingWaitlistStatus(input.Status), Valid: input.Status != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create waitlist entry", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to retrieve last insert ID for waitlist entry",
			logger.Any("owner_id", ownerID),
			logger.Error(err),
		)
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit waitlist transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.WaitlistInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin waitlist update transaction", logger.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, ownerID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
		Phone: input.CustomerPhone,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync waitlist customer", logger.Error(err))
		return err
	}

	err = r.queries.WithTx(tx).UpdateWaitlistEntry(ctx, sqlc.UpdateWaitlistEntryParams{
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		PreferredDate: parseNullDate(input.PreferredDate),
		PreferredTime: parseNullTime(input.PreferredTime),
		Status:        sqlc.NullBookingWaitlistStatus{BookingWaitlistStatus: sqlc.BookingWaitlistStatus(input.Status), Valid: input.Status != ""},
		OwnerID:       ownerID,
		ID:            id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update waitlist entry", logger.Error(err))
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit waitlist update transaction", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.BookingWaitlist) domain.WaitlistEntry {
	w := domain.WaitlistEntry{ID: row.ID, OwnerID: row.OwnerID}
	if row.ServiceID.Valid {
		w.ServiceID = row.ServiceID.Int64
	}
	if row.CustomerName.Valid {
		w.CustomerName = row.CustomerName.String
	}
	if row.CustomerEmail.Valid {
		w.CustomerEmail = row.CustomerEmail.String
	}
	if row.CustomerPhone.Valid {
		w.CustomerPhone = row.CustomerPhone.String
	}
	if row.PreferredDate.Valid {
		w.PreferredDate = row.PreferredDate.Time.Format("2006-01-02")
	}
	if row.PreferredTime.Valid {
		w.PreferredTime = row.PreferredTime.Time.Format("15:04:05")
	}
	if row.Status.Valid {
		w.Status = string(row.Status.BookingWaitlistStatus)
	}
	if row.CreatedAt.Valid {
		w.CreatedAt = row.CreatedAt.Time
	}
	return w
}

func parseNullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		logger.Global().Warn("failed to parse waitlist preferred date value",
			logger.String("value", s),
			logger.Error(err),
		)
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func parseNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		logger.Global().Warn("failed to parse waitlist preferred time value",
			logger.String("value", s),
			logger.Error(err),
		)
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
