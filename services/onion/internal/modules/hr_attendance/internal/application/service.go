package application

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"sync"
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

func (s *Service) ClockIn(ctx context.Context, businessID int64, record domain.AttendanceRecordRecord) (*domain.AttendanceRecordRecord, error) {
	if !domain.IsValidClockInMethod(record.ClockInMethod) {
		return nil, domain.ErrInvalidClockInMethod
	}

	cfg, err := s.repo.GetAttendanceConfig(ctx, businessID)
	if err != nil {
		return nil, err
	}

	if cfg != nil {
		if err := s.validateClockInMethod(cfg, record.ClockInMethod); err != nil {
			return nil, err
		}
	}

	active, err := s.repo.GetActiveClockIn(ctx, businessID, record.EmployeeID)
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
			loc, locErr := s.repo.GetLocationCoordinates(ctx, businessID, record.LocationID)
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

	// UTC everywhere else in the HR modules (payroll periods, report
	// ranges) — the attendance day bucket must agree with them regardless
	// of the host's local timezone.
	if record.ClockInAt.IsZero() {
		record.ClockInAt = time.Now().UTC()
	}
	if record.Date == "" {
		record.Date = record.ClockInAt.UTC().Format("2006-01-02")
	}
	if record.Status == "" {
		record.Status = "present"
	}

	return s.repo.CreateAttendanceRecord(ctx, businessID, record)
}

func (s *Service) ClockOut(ctx context.Context, ownerID, employeeID int64) (*domain.AttendanceRecordRecord, error) {
	active, err := s.repo.GetActiveClockIn(ctx, ownerID, employeeID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, domain.ErrNotClockedIn
	}

	now := time.Now().UTC()
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
	qrCache.Put(ctx, key, token, cacher.WithExpiry(ctx, 5*time.Minute))

	return token, nil
}

func (s *Service) ValidateQRToken(ctx context.Context, ownerID, locationID int64, token string) error {
	key := fmt.Sprintf("hr_attendance:%d:qr:%d", ownerID, locationID)

	qrCache := cacher.Of[string]()
	cached, found := qrCache.GetAndDelete(ctx, key)
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

// PIN lockout policy (#817): a typically-4-digit PIN brute-forced through the
// anonymous kiosk endpoint otherwise sees only the global rate limit. Five
// failures lock the employee's PIN for 15 minutes, tracked in Redis when
// available. Without Redis, lockout is best-effort and depends on the
// deployment's cache configuration.
const (
	pinMaxFailures = 5
	pinLockoutTTL  = 15 * time.Minute
)

type pinFailureState struct {
	count     int
	expiresAt time.Time
}

var pinFailureFallback = struct {
	sync.Mutex
	entries map[string]pinFailureState
}{
	entries: make(map[string]pinFailureState),
}

func (s *Service) pinFailureKey(ownerID, employeeID int64) string {
	return fmt.Sprintf("hr_attendance:pinfail:%d:%d", ownerID, employeeID)
}

// IsPINLocked reports whether the employee's PIN is currently locked out.
func (s *Service) IsPINLocked(ctx context.Context, ownerID, employeeID int64) bool {
	key := s.pinFailureKey(ownerID, employeeID)
	cache := cacher.GetRedisCache()
	if cache != nil {
		raw := cache.Get(ctx, key)
		if len(raw) > 0 {
			count, err := strconv.Atoi(string(raw))
			if err == nil {
				return count >= pinMaxFailures
			}
		}
	}
	return localPINFailureCount(key) >= pinMaxFailures
}

// RegisterPINFailure records a wrong-PIN attempt.
func (s *Service) RegisterPINFailure(ctx context.Context, ownerID, employeeID int64) {
	key := s.pinFailureKey(ownerID, employeeID)
	cache := cacher.GetRedisCache()
	if cache != nil {
		// Increment is atomic in Redis; a read/modify/write sequence would let
		// concurrent guesses overwrite one another and bypass the five-attempt cap.
		if count := cache.Increment(ctx, key, pinLockoutTTL); count > 0 {
			return
		}
	}
	// The no-op cache is used when Redis is disabled, and Increment returns 0
	// when Redis is unavailable. Keep the lockout effective in either case.
	incrementLocalPINFailure(key)
}

// ClearPINFailures resets the counter after a successful validation.
func (s *Service) ClearPINFailures(ctx context.Context, ownerID, employeeID int64) {
	key := s.pinFailureKey(ownerID, employeeID)
	cache := cacher.GetRedisCache()
	if cache != nil {
		cache.Delete(ctx, []string{key})
	}
	clearLocalPINFailures(key)
}

func localPINFailureCount(key string) int {
	now := time.Now()
	pinFailureFallback.Lock()
	defer pinFailureFallback.Unlock()

	entry, ok := pinFailureFallback.entries[key]
	if !ok || !now.Before(entry.expiresAt) {
		delete(pinFailureFallback.entries, key)
		return 0
	}
	return entry.count
}

func incrementLocalPINFailure(key string) int {
	now := time.Now()
	pinFailureFallback.Lock()
	defer pinFailureFallback.Unlock()

	entry := pinFailureFallback.entries[key]
	if !now.Before(entry.expiresAt) {
		entry = pinFailureState{}
	}
	entry.count++
	entry.expiresAt = now.Add(pinLockoutTTL)
	pinFailureFallback.entries[key] = entry
	return entry.count
}

func clearLocalPINFailures(key string) {
	pinFailureFallback.Lock()
	delete(pinFailureFallback.entries, key)
	pinFailureFallback.Unlock()
}

func (s *Service) ValidatePIN(ctx context.Context, ownerID, employeeID int64, pin string) error {
	hash, err := s.repo.GetEmployeePINHash(ctx, ownerID, employeeID)
	if err != nil {
		return err
	}
	if !passwordutils.CheckPasswordHash(pin, hash) {
		s.RegisterPINFailure(ctx, ownerID, employeeID)
		return domain.ErrInvalidPIN
	}
	s.ClearPINFailures(ctx, ownerID, employeeID)
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
