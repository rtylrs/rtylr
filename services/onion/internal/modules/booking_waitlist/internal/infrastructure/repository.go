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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.WaitlistEntry, error) {
	rows, err := r.queries.ListWaitlist(ctx, businessID)
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
// owned by businessID. It returns domain.ErrServiceNotFound when no matching row is
// found, so a caller cannot reference another tenant's service.
func (r *Repository) ServiceBelongsToOwner(ctx context.Context, businessID, serviceID int64) error {
	// cross-schema read from booking_services — verify service ownership before linking a waitlist entry
	_, err := r.queries.GetBookingServiceForOwner(ctx, sqlc.GetBookingServiceForOwnerParams{
		ID:         serviceID,
		BusinessID: businessID,
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

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.WaitlistInput) (int64, error) {
	preferredDate, err := parseNullDate(input.PreferredDate)
	if err != nil {
		return 0, err
	}
	preferredTime, err := parseNullTime(input.PreferredTime)
	if err != nil {
		return 0, err
	}

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

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
		Phone: input.CustomerPhone,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync waitlist customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateWaitlistEntry(ctx, sqlc.CreateWaitlistEntryParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		PreferredDate: preferredDate,
		PreferredTime: preferredTime,
		Status:        sqlc.NullBookingWaitlistStatus{BookingWaitlistStatus: sqlc.BookingWaitlistStatus(input.Status), Valid: input.Status != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create waitlist entry", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to retrieve last insert ID for waitlist entry",
			logger.Any("business_id", businessID),
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

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.WaitlistInput) error {
	preferredDate, err := parseNullDate(input.PreferredDate)
	if err != nil {
		return err
	}
	preferredTime, err := parseNullTime(input.PreferredTime)
	if err != nil {
		return err
	}

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

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
		Phone: input.CustomerPhone,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync waitlist customer", logger.Error(err))
		return err
	}

	result, err := r.queries.WithTx(tx).UpdateWaitlistEntry(ctx, sqlc.UpdateWaitlistEntryParams{
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		PreferredDate: preferredDate,
		PreferredTime: preferredTime,
		Status:        sqlc.NullBookingWaitlistStatus{BookingWaitlistStatus: sqlc.BookingWaitlistStatus(input.Status), Valid: input.Status != ""},
		BusinessID:    businessID,
		ID:            id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update waitlist entry", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read affected rows for waitlist update", logger.Error(err))
		return err
	}
	if affected == 0 {
		// MySQL reports changed rows, not matched rows, so re-sending the
		// stored values also lands here. Read the row back to tell that
		// idempotent no-op apart from a missing (or another tenant's) entry.
		if _, err = r.queries.WithTx(tx).GetWaitlistEntry(ctx, sqlc.GetWaitlistEntryParams{
			BusinessID: businessID,
			ID:         id,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrWaitlistEntryNotFound
			}
			logger.Ctx(ctx).Error("Failed to verify waitlist entry after update", logger.Error(err))
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit waitlist update transaction", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.BookingWaitlist) domain.WaitlistEntry {
	w := domain.WaitlistEntry{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID}
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
		w.PreferredTime = row.PreferredTime.String
	}
	if row.Status.Valid {
		w.Status = string(row.Status.BookingWaitlistStatus)
	}
	if row.CreatedAt.Valid {
		w.CreatedAt = row.CreatedAt.Time
	}
	return w
}

// parseNullDate maps an optional YYYY-MM-DD value onto sql.NullTime. An empty
// string means "not provided" (NULL); anything else that does not parse is an
// error, because persisting NULL would erase a real client value while the API
// still reports success.
func parseNullDate(s string) (sql.NullTime, error) {
	if s == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}, domain.ErrPreferredDateInvalid
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

// parseNullTime maps an optional 24-hour time onto sql.NullString. Both HH:MM
// and HH:MM:SS are accepted and stored normalised to HH:MM:SS.
//
// preferred_time is a MySQL TIME column: the driver returns TIME as raw bytes,
// so scanning it into time.Time fails with "unsupported Scan" (see the
// sqlc.yaml override), and time.Parse of a clock-only layout produces year 0,
// which the driver refuses to write. Keeping the value a string end to end
// avoids both traps.
func parseNullTime(s string) (sql.NullString, error) {
	if s == "" {
		return sql.NullString{}, nil
	}
	layout := "15:04:05"
	if len(s) == len("15:04") {
		layout = "15:04"
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return sql.NullString{}, domain.ErrPreferredTimeInvalid
	}
	return sql.NullString{String: t.Format("15:04:05"), Valid: true}, nil
}
