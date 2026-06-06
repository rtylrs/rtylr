package application

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) ClockIn(ctx context.Context, ownerID int64, record domain.AttendanceRecordRecord) (*domain.AttendanceRecordRecord, error) {
	if !domain.IsValidClockInMethod(record.ClockInMethod) {
		return nil, domain.ErrInvalidClockInMethod
	}

	cfg, err := s.repo.GetAttendanceConfig(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	if cfg != nil {
		if err := s.validateClockInMethod(cfg, record.ClockInMethod); err != nil {
			return nil, err
		}
	}

	active, err := s.repo.GetActiveClockIn(ctx, ownerID, record.EmployeeID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, domain.ErrAlreadyClockedIn
	}

	if cfg != nil && cfg.GPSEnabled && record.ClockInMethod == "gps" {
		refLat := cfg.DefaultLocationLat
		refLng := cfg.DefaultLocationLng

		if record.LocationID != 0 {
			loc, locErr := s.repo.GetLocationCoordinates(ctx, ownerID, record.LocationID)
			if locErr != nil {
				return nil, locErr
			}
			if loc != nil && loc.HasCoords {
				refLat = loc.Latitude
				refLng = loc.Longitude
			}
		}

		dist := haversineDistance(refLat, refLng, record.ClockInLat, record.ClockInLng)
		if dist > float64(cfg.GeofenceRadiusM) {
			return nil, domain.ErrGeofenceViolation
		}
	}

	if record.ClockInAt.IsZero() {
		record.ClockInAt = time.Now()
	}
	if record.Date == "" {
		record.Date = record.ClockInAt.Format("2006-01-02")
	}
	if record.Status == "" {
		record.Status = "present"
	}

	return s.repo.CreateAttendanceRecord(ctx, ownerID, record)
}

func (s *Service) ClockOut(ctx context.Context, ownerID, employeeID int64) (*domain.AttendanceRecordRecord, error) {
	active, err := s.repo.GetActiveClockIn(ctx, ownerID, employeeID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, domain.ErrNotClockedIn
	}

	now := time.Now()
	duration := int32(now.Sub(active.ClockInAt).Minutes())

	active.ClockOutAt = now
	active.DurationMinutes = duration
	active.Status = "present"

	return s.repo.UpdateAttendanceRecord(ctx, ownerID, *active)
}

func (s *Service) GetMyAttendance(ctx context.Context, ownerID, employeeID int64, params domain.ListAttendanceParams) ([]domain.AttendanceRecordRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	params.EmployeeID = employeeID
	return s.repo.ListAttendance(ctx, ownerID, params)
}

func (s *Service) GetLocationAttendance(ctx context.Context, ownerID, locationID int64, date string) ([]domain.AttendanceRecordRecord, error) {
	return s.repo.GetDailyAttendance(ctx, ownerID, locationID, date)
}

func (s *Service) GenerateQRToken(ctx context.Context, ownerID, locationID int64) (string, error) {
	token := uuid.New().String()
	key := fmt.Sprintf("hr_attendance:%d:qr:%d", ownerID, locationID)

	qrCache := cacher.Of[string]()
	qrCache.Put(ctx, key, token, cacher.WithExpiry(ctx, 24*time.Hour))

	return token, nil
}

func (s *Service) ValidateQRToken(ctx context.Context, ownerID, locationID int64, token string) error {
	key := fmt.Sprintf("hr_attendance:%d:qr:%d", ownerID, locationID)

	qrCache := cacher.Of[string]()
	cached, found := qrCache.Get(ctx, key)
	if !found || cached != token {
		return domain.ErrInvalidQRToken
	}
	return nil
}

func (s *Service) SetClockInPIN(ctx context.Context, ownerID, employeeID int64, pin string) error {
	hash, err := passwordutils.HashPassword(pin)
	if err != nil {
		return err
	}
	return s.repo.SetEmployeePINHash(ctx, ownerID, employeeID, hash)
}

func (s *Service) ValidatePIN(ctx context.Context, ownerID, employeeID int64, pin string) error {
	hash, err := s.repo.GetEmployeePINHash(ctx, ownerID, employeeID)
	if err != nil {
		return err
	}
	if !passwordutils.CheckPasswordHash(pin, hash) {
		return domain.ErrInvalidPIN
	}
	return nil
}

func (s *Service) validateClockInMethod(cfg *domain.AttendanceConfigRecord, method string) error {
	switch method {
	case domain.ClockInMethodApp:
		if !cfg.AppEnabled {
			return domain.ErrClockInMethodDisabled
		}
	case domain.ClockInMethodQR:
		if !cfg.QREnabled {
			return domain.ErrClockInMethodDisabled
		}
	case domain.ClockInMethodGPS:
		if !cfg.GPSEnabled {
			return domain.ErrClockInMethodDisabled
		}
	case domain.ClockInMethodPin:
		if !cfg.PinEnabled {
			return domain.ErrClockInMethodDisabled
		}
	case domain.ClockInMethodFacial:
		if !cfg.FacialEnabled {
			return domain.ErrClockInMethodDisabled
		}
	default:
		return domain.ErrInvalidClockInMethod
	}
	return nil
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}
