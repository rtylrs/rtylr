package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Appointment, error) {
	rows, err := r.queries.ListAppointments(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list appointments", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Appointment, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) ListToday(ctx context.Context, businessID int64) ([]domain.Appointment, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	rows, err := r.queries.ListAppointmentsByDateRange(ctx, sqlc.ListAppointmentsByDateRangeParams{
		BusinessID:  businessID,
		StartTime:   startOfDay,
		StartTime_2: endOfDay,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list today appointments", logger.Error(err))
		return nil, err
	}
	items := make([]domain.Appointment, 0, len(rows))
	for _, row := range rows {
		items = append(items, fromRow(row))
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.Appointment, error) {
	row, err := r.queries.GetAppointment(ctx, sqlc.GetAppointmentParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get appointment", logger.Error(err))
		return domain.Appointment{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.AppointmentInput) (int64, error) {
	startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", input.StartTime)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse start_time", logger.Error(err))
		return 0, err
	}
	endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", input.EndTime)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse end_time", logger.Error(err))
		return 0, err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin appointment transaction", logger.Error(err))
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
		logger.Ctx(ctx).Error("Failed to sync appointment customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateAppointment(ctx, sqlc.CreateAppointmentParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		ResourceID:    sql.NullInt64{Int64: input.ResourceID, Valid: input.ResourceID != 0},
		StaffID:       sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		StartTime:     startTime, EndTime: endTime,
		Status: sqlc.NullBookingAppointmentsStatus{BookingAppointmentsStatus: sqlc.BookingAppointmentsStatus(input.Status), Valid: input.Status != ""},
		Notes:  sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create appointment", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to retrieve last insert ID for appointment", logger.Error(err))
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit appointment transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.AppointmentInput) error {
	startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", input.StartTime)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse start_time", logger.Error(err))
		return err
	}
	endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", input.EndTime)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse end_time", logger.Error(err))
		return err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin appointment update transaction", logger.Error(err))
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
		logger.Ctx(ctx).Error("Failed to sync appointment customer", logger.Error(err))
		return err
	}

	err = r.queries.WithTx(tx).UpdateAppointment(ctx, sqlc.UpdateAppointmentParams{
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		ResourceID:    sql.NullInt64{Int64: input.ResourceID, Valid: input.ResourceID != 0},
		StaffID:       sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		StartTime:     startTime, EndTime: endTime,
		Status:     sqlc.NullBookingAppointmentsStatus{BookingAppointmentsStatus: sqlc.BookingAppointmentsStatus(input.Status), Valid: input.Status != ""},
		Notes:      sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		BusinessID: businessID, ID: id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update appointment", logger.Error(err))
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit appointment update transaction", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.DeleteAppointment(ctx, sqlc.DeleteAppointmentParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete appointment", logger.Error(err))
		return err
	}
	return nil
}

func fromRow(row sqlc.BookingAppointment) domain.Appointment {
	a := domain.Appointment{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, StartTime: row.StartTime, EndTime: row.EndTime}
	if row.ServiceID.Valid {
		a.ServiceID = row.ServiceID.Int64
	}
	if row.ResourceID.Valid {
		a.ResourceID = row.ResourceID.Int64
	}
	if row.StaffID.Valid {
		a.StaffID = row.StaffID.Int64
	}
	if row.CustomerName.Valid {
		a.CustomerName = row.CustomerName.String
	}
	if row.CustomerEmail.Valid {
		a.CustomerEmail = row.CustomerEmail.String
	}
	if row.CustomerPhone.Valid {
		a.CustomerPhone = row.CustomerPhone.String
	}
	if row.Status.Valid {
		a.Status = string(row.Status.BookingAppointmentsStatus)
	}
	if row.Notes.Valid {
		a.Notes = row.Notes.String
	}
	if row.CreatedAt.Valid {
		a.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		a.UpdatedAt = row.UpdatedAt.Time
	}
	return a
}
