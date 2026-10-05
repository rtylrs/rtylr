package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/voxire/onion/internal/modules/booking/internal/outbox"
	"strings"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/booking/appointments/internal/domain"
	"github.com/voxire/onion/internal/modules/booking/appointments/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
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
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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
	// Rollback after Commit returns ErrTxDone and is a no-op, so it is safe to
	// defer unconditionally: every early return must release the transaction,
	// or the tenant FOR UPDATE lock taken below is held until the request
	// context is cancelled.
	defer func() { _ = tx.Rollback() }()
	qtx := r.queries.WithTx(tx)
	status := normalizeAppointmentStatus(input.Status)
	if status == "" {
		status = "pending"
	}
	if !validAppointmentStatus(status) {
		return 0, domain.ErrInvalidStatus
	}
	if err = qtx.EnsureScheduleLock(ctx, businessID); err != nil {
		return 0, err
	}
	if _, err = qtx.LockBusinessForBooking(ctx, businessID); err != nil {
		return 0, err
	}
	if err = ensureAppointmentReferences(ctx, qtx, businessID, input.ServiceID, input.ResourceID); err != nil {
		return 0, err
	}
	if status != "cancelled" && status != "no_show" {
		if err = ensureNoAppointmentOverlap(ctx, qtx, businessID, startTime, endTime, 0, input.ResourceID, input.StaffID); err != nil {
			return 0, err
		}
	}

	result, err := qtx.CreateAppointment(ctx, sqlc.CreateAppointmentParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		ResourceID:    sql.NullInt64{Int64: input.ResourceID, Valid: input.ResourceID != 0},
		StaffID:       sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		StartTime:     startTime, EndTime: endTime,
		Status: sqlc.NullBookingAppointmentsStatus{BookingAppointmentsStatus: sqlc.BookingAppointmentsStatus(status), Valid: true},
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
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail, Phone: input.CustomerPhone}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "booking_appointment",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("booking_appointment", id, contact),
	}); err != nil {
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
	// Rollback after Commit returns ErrTxDone and is a no-op, so it is safe to
	// defer unconditionally: sentinel returns (invalid status, missing row)
	// must not abandon the transaction holding the tenant FOR UPDATE lock.
	defer func() { _ = tx.Rollback() }()
	qtx := r.queries.WithTx(tx)
	if err = qtx.EnsureScheduleLock(ctx, businessID); err != nil {
		return err
	}
	if _, err = qtx.LockBusinessForBooking(ctx, businessID); err != nil {
		return err
	}
	current, err := qtx.GetAppointmentStatusForUpdate(ctx, sqlc.GetAppointmentStatusForUpdateParams{BusinessID: businessID, ID: id})
	if err != nil {
		return err
	}
	currentStatus := normalizeAppointmentStatus(string(current.BookingAppointmentsStatus))
	if currentStatus == "" {
		currentStatus = "pending"
	}
	status := normalizeAppointmentStatus(input.Status)
	if status == "" {
		status = currentStatus
	}
	if !validAppointmentStatus(status) || !validAppointmentTransition(currentStatus, status) {
		return domain.ErrInvalidStatus
	}
	if err = ensureAppointmentReferences(ctx, qtx, businessID, input.ServiceID, input.ResourceID); err != nil {
		return err
	}
	if status != "cancelled" && status != "no_show" {
		if err = ensureNoAppointmentOverlap(ctx, qtx, businessID, startTime, endTime, id, input.ResourceID, input.StaffID); err != nil {
			return err
		}
	}

	// Existence is already proven by the locked status read above, so a zero
	// RowsAffected (no column actually changed) must not be treated as a
	// missing appointment: the driver does not use clientFoundRows, and
	// updated_at is only bumped by MySQL when another column changes.
	if _, err = qtx.UpdateAppointment(ctx, sqlc.UpdateAppointmentParams{
		ServiceID:     sql.NullInt64{Int64: input.ServiceID, Valid: input.ServiceID != 0},
		ResourceID:    sql.NullInt64{Int64: input.ResourceID, Valid: input.ResourceID != 0},
		StaffID:       sql.NullInt64{Int64: input.StaffID, Valid: input.StaffID != 0},
		CustomerName:  sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerPhone: sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		StartTime:     startTime, EndTime: endTime,
		Status:     sqlc.NullBookingAppointmentsStatus{BookingAppointmentsStatus: sqlc.BookingAppointmentsStatus(status), Valid: true},
		Notes:      sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		BusinessID: businessID, ID: id,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update appointment", logger.Error(err))
		return err
	}
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail, Phone: input.CustomerPhone}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "booking_appointment",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("booking_appointment", id, contact) + ":" + time.Now().UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit appointment update transaction", logger.Error(err))
		return err
	}
	return nil
}

func ensureNoAppointmentOverlap(ctx context.Context, qtx *sqlc.Queries, businessID int64, startTime, endTime time.Time, excludeID, resourceID, staffID int64) error {
	if resourceID <= 0 && staffID <= 0 {
		return nil
	}
	rows, err := qtx.LockAppointmentOverlaps(ctx, sqlc.LockAppointmentOverlapsParams{
		BusinessID: businessID,
		// sqlc binds these fields to the placeholders in the overlap predicates:
		// start_time < ? receives the requested end, and end_time > ? receives
		// the requested start.
		StartTime:  endTime,
		EndTime:    startTime,
		Column4:    excludeID,
		ID:         excludeID,
		Column6:    resourceID,
		ResourceID: nullableID(resourceID),
		Column8:    staffID,
		StaffID:    nullableID(staffID),
	})
	if err != nil {
		return fmt.Errorf("check appointment availability: %w", err)
	}
	if len(rows) > 0 {
		return domain.ErrTimeSlotUnavailable
	}
	return nil
}

func ensureAppointmentReferences(ctx context.Context, qtx *sqlc.Queries, businessID, serviceID, resourceID int64) error {
	if serviceID > 0 {
		if _, err := qtx.GetBookingServiceForAppointment(ctx, sqlc.GetBookingServiceForAppointmentParams{ID: serviceID, BusinessID: businessID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrInvalidReference
			}
			return err
		}
	}
	if resourceID > 0 {
		if _, err := qtx.GetBookingResourceForAppointment(ctx, sqlc.GetBookingResourceForAppointmentParams{ID: resourceID, BusinessID: businessID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrInvalidReference
			}
			return err
		}
	}
	return nil
}

func nullableID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}

func normalizeAppointmentStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func validAppointmentStatus(status string) bool {
	switch status {
	case "pending", "confirmed", "in_progress", "completed", "cancelled", "no_show":
		return true
	default:
		return false
	}
}

func validAppointmentTransition(current, next string) bool {
	if current == next {
		return true
	}
	switch current {
	case "pending":
		return next == "confirmed" || next == "cancelled" || next == "no_show"
	case "confirmed":
		return next == "in_progress" || next == "cancelled" || next == "no_show"
	case "in_progress":
		return next == "completed" || next == "cancelled" || next == "no_show"
	default:
		return false
	}
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteAppointment(ctx, sqlc.DeleteAppointmentParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete appointment", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
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
