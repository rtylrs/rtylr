package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) CreateAttendanceRecord(ctx context.Context, businessID int64, record domain.AttendanceRecordRecord) (*domain.AttendanceRecordRecord, error) {
	date, err := time.Parse("2006-01-02", record.Date)
	if err != nil {
		return nil, err
	}

	clockInAt := record.ClockInAt
	if clockInAt.IsZero() {
		clockInAt = time.Now()
	}

	ownerID := record.OwnerID
	if ownerID == 0 {
		ownerID = businessID
	}

	result, err := r.queries.CreateAttendanceRecord(ctx, sqlc.CreateAttendanceRecordParams{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		EmployeeID:    record.EmployeeID,
		LocationID:    record.LocationID,
		Date:          date,
		ClockInAt:     clockInAt,
		ClockInMethod: record.ClockInMethod,
		ClockInLat:    nullFloat64(record.ClockInLat),
		ClockInLng:    nullFloat64(record.ClockInLng),
		IsLate:        record.IsLate,
		LateMinutes:   record.LateMinutes,
		Status:        record.Status,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create attendance record", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	record.ID = id
	record.OwnerID = ownerID
	record.BusinessID = businessID
	return &record, nil
}

func (r *Repository) UpdateAttendanceRecord(ctx context.Context, businessID int64, record domain.AttendanceRecordRecord) (*domain.AttendanceRecordRecord, error) {
	err := r.queries.UpdateAttendanceRecord(ctx, sqlc.UpdateAttendanceRecordParams{
		ClockOutAt:      sql.NullTime{Time: record.ClockOutAt, Valid: !record.ClockOutAt.IsZero()},
		DurationMinutes: sql.NullInt32{Int32: record.DurationMinutes, Valid: record.DurationMinutes > 0},
		Status:          record.Status,
		ID:              record.ID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update attendance record", logger.Any("business_id", businessID), logger.Any("record_id", record.ID), logger.Error(err))
		return nil, err
	}
	record.BusinessID = businessID
	return &record, nil
}

func (r *Repository) GetActiveClockIn(ctx context.Context, businessID, employeeID int64) (*domain.AttendanceRecordRecord, error) {
	row, err := r.queries.GetActiveClockIn(ctx, sqlc.GetActiveClockInParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	rec := attendanceFromRow(row)
	return &rec, nil
}

func (r *Repository) ListAttendance(ctx context.Context, businessID int64, params domain.ListAttendanceParams) ([]domain.AttendanceRecordRecord, int64, error) {
	dateFrom, err := time.Parse("2006-01-02", params.DateFrom)
	if err != nil {
		return nil, 0, err
	}
	dateTo, err := time.Parse("2006-01-02", params.DateTo)
	if err != nil {
		return nil, 0, err
	}

	// When EmployeeID is 0, query all employees for the business.
	if params.EmployeeID == 0 {
		return r.listAllAttendance(ctx, businessID, dateFrom, dateTo, params.Limit, params.Offset)
	}

	total, err := r.queries.CountAttendance(ctx, sqlc.CountAttendanceParams{
		BusinessID: businessID,
		EmployeeID: params.EmployeeID,
		Date:       dateFrom,
		Date_2:     dateTo,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count attendance", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListAttendance(ctx, sqlc.ListAttendanceParams{
		BusinessID: businessID,
		EmployeeID: params.EmployeeID,
		Date:       dateFrom,
		Date_2:     dateTo,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list attendance", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	records := make([]domain.AttendanceRecordRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, attendanceFromRow(row))
	}
	return records, total, nil
}

func (r *Repository) listAllAttendance(ctx context.Context, businessID int64, dateFrom, dateTo time.Time, limit, offset int32) ([]domain.AttendanceRecordRecord, int64, error) {
	total, err := r.queries.CountAllAttendance(ctx, sqlc.CountAllAttendanceParams{
		BusinessID: businessID,
		Date:       dateFrom,
		Date_2:     dateTo,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListAllAttendance(ctx, sqlc.ListAllAttendanceParams{
		BusinessID: businessID,
		Date:       dateFrom,
		Date_2:     dateTo,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.AttendanceRecordRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, attendanceFromRow(row))
	}
	return records, total, nil
}

func (r *Repository) GetDailyAttendance(ctx context.Context, businessID, locationID int64, date string) ([]domain.AttendanceRecordRecord, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetDailyAttendance(ctx, sqlc.GetDailyAttendanceParams{
		BusinessID: businessID,
		LocationID: locationID,
		Date:       d,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get daily attendance", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	records := make([]domain.AttendanceRecordRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, attendanceFromRow(row))
	}
	return records, nil
}

func (r *Repository) GetAttendanceConfig(ctx context.Context, businessID int64) (*domain.AttendanceConfigRecord, error) {
	row, err := r.queries.GetAttendanceConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Ctx(ctx).Error("failed to get attendance config", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	return &domain.AttendanceConfigRecord{
		ID:                 row.ID,
		OwnerID:            row.OwnerID,
		BusinessID:         row.BusinessID,
		AppEnabled:         row.AppEnabled,
		QREnabled:          row.QrEnabled,
		GPSEnabled:         row.GpsEnabled,
		PinEnabled:         row.PinEnabled,
		FacialEnabled:      row.FacialEnabled,
		GeofenceRadiusM:    row.GeofenceRadiusM,
		DefaultLocationLat: row.DefaultLocationLat,
		DefaultLocationLng: row.DefaultLocationLng,
	}, nil
}

// --- Employee PIN ---

func (r *Repository) GetEmployeePINHash(ctx context.Context, businessID, employeeID int64) (string, error) {
	// cross-schema read from hr_core — employee PIN hash for attendance clock-in verification
	hash, err := r.queries.GetEmployeePINHash(ctx, sqlc.GetEmployeePINHashParams{
		ID: employeeID, BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrAttendanceNotFound
		}
		logger.Ctx(ctx).Error("failed to get employee PIN hash", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return "", err
	}
	if !hash.Valid || hash.String == "" {
		return "", domain.ErrPINNotSet
	}
	return hash.String, nil
}

func (r *Repository) SetEmployeePINHash(ctx context.Context, businessID, employeeID int64, hash string) error {
	// cross-schema write to hr_core — set employee PIN hash for attendance clock-in
	result, err := r.queries.SetEmployeePINHash(ctx, sqlc.SetEmployeePINHashParams{
		ClockInPinHash: sql.NullString{String: hash, Valid: hash != ""},
		ID:             employeeID,
		BusinessID:     businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to set employee PIN hash", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAttendanceNotFound
	}
	return nil
}

// --- Location Coordinates ---

func (r *Repository) GetLocationCoordinates(ctx context.Context, businessID, locationID int64) (*domain.LocationCoordinates, error) {
	// cross-schema read from hr_core — location coordinates for geofence attendance validation
	row, err := r.queries.GetLocationCoordinates(ctx, sqlc.GetLocationCoordinatesParams{
		ID: locationID, BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Ctx(ctx).Error("failed to get location coordinates", logger.Any("business_id", businessID), logger.Any("location_id", locationID), logger.Error(err))
		return nil, err
	}
	return &domain.LocationCoordinates{
		Latitude:  row.Latitude.Float64,
		Longitude: row.Longitude.Float64,
		HasCoords: row.Latitude.Valid && row.Longitude.Valid && (row.Latitude.Float64 != 0 || row.Longitude.Float64 != 0),
	}, nil
}

// --- Helpers ---

func attendanceFromRow(row sqlc.HrAttendanceRecord) domain.AttendanceRecordRecord {
	return domain.AttendanceRecordRecord{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		EmployeeID:      row.EmployeeID,
		LocationID:      row.LocationID,
		Date:            row.Date.Format("2006-01-02"),
		ClockInAt:       row.ClockInAt,
		ClockOutAt:      row.ClockOutAt.Time,
		DurationMinutes: nullInt32Value(row.DurationMinutes),
		ClockInMethod:   row.ClockInMethod,
		ClockInLat:      row.ClockInLat.Float64,
		ClockInLng:      row.ClockInLng.Float64,
		IsLate:          row.IsLate,
		LateMinutes:     row.LateMinutes,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func nullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: math.Abs(v) > 0.000001}
}

func nullInt32Value(v sql.NullInt32) int32 {
	if v.Valid {
		return v.Int32
	}
	return 0
}
