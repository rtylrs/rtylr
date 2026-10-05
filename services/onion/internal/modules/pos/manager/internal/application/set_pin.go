package application

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// managerPINWriter is the slice of the repository SetPIN needs: the credential
// write. Narrowing it keeps the set path testable without a database and stops
// it growing new reads by accident.
type managerPINWriter interface {
	SetManagerPIN(ctx context.Context, input domain.SetPINInput, pinHash string, events ...audit.Event) error
}

// ownerCredentialVerifier checks the account password without exposing its hash.
type ownerCredentialVerifier interface {
	VerifyAccountPassword(ctx context.Context, userID int64, password string) (bool, error)
}

// SetPIN requires the owner's account password and the durable per-business
// guess budget before minting an override credential. Missing dependencies fail closed.
func (s *Service) SetPIN(ctx context.Context, input domain.SetPINInput) error {
	if input.BusinessID <= 0 || input.StaffID <= 0 || input.OwnerID <= 0 {
		return domain.ErrInvalidSetPIN
	}
	pin := strings.TrimSpace(input.PIN)
	if pin == "" || strings.TrimSpace(input.CurrentPassword) == "" {
		return domain.ErrInvalidSetPIN
	}
	// The policy is checked before any lookup or credential work. A weak PIN is
	// the caller's own input, so rejecting it early touches no secret, spends no
	// guess budget, and tells the operator which rule they broke.
	if err := domain.ValidatePIN(pin); err != nil {
		return err
	}

	writer := s.writer
	credentials := s.credentials
	store := s.attempts
	if writer == nil || credentials == nil || store == nil {
		// Fail closed: without the write path or the counter this request cannot
		// be stored, and without the counter it cannot be rate limited.
		logger.Ctx(ctx).Error("manager PIN set path is not wired; denying",
			logger.Any("business_id", input.BusinessID),
		)
		passwordutils.CheckDummyHash(input.CurrentPassword)
		return domain.ErrSetPINStoreUnavailable
	}

	// The same durable per-business budget that brakes PIN guessing brakes
	// owner-password guessing here. An unreadable counter, an active backoff and
	// a wrong password are answered identically so the caller cannot use this
	// endpoint to probe the business's lockout state.
	retryAt, err := store.PinRetryAt(ctx, input.BusinessID)
	if err != nil {
		logger.Ctx(ctx).Error("manager pin attempt counter unavailable; denying PIN change",
			logger.Any("business_id", input.BusinessID),
			logger.Error(err),
		)
		passwordutils.CheckDummyHash(input.CurrentPassword)
		return domain.ErrSetPINNotAuthorized
	}
	if !retryAt.IsZero() && s.clock().Before(retryAt) {
		passwordutils.CheckDummyHash(input.CurrentPassword)
		return domain.ErrSetPINNotAuthorized
	}

	valid, err := credentials.VerifyAccountPassword(ctx, input.OwnerID, input.CurrentPassword)
	if err != nil {
		logger.Ctx(ctx).Error("failed to verify owner credential for manager PIN change",
			logger.Any("business_id", input.BusinessID), logger.Error(err))
		return domain.ErrSetPINStoreUnavailable
	}
	if !valid {
		s.recordPINFailure(ctx, store, input.BusinessID)
		return domain.ErrSetPINNotAuthorized
	}

	pinHash, err := passwordutils.HashPIN(pin)
	if err != nil {
		logger.Ctx(ctx).Error("failed to hash manager PIN", logger.Error(err))
		return err
	}
	// Drop the plaintext from the struct before it can reach a log line or an
	// error report further down the call stack.
	pin = ""
	input.PIN = ""

	if err := writer.SetManagerPIN(ctx, input, pinHash, audit.Event{
		BusinessID:  input.BusinessID,
		ActorID:     input.OwnerID,
		EntityType:  "pos_manager_pin",
		EntityID:    input.StaffID,
		Action:      "set_pin",
		Description: "Manager override PIN set",
	}); err != nil {
		logger.Ctx(ctx).Error("failed to store manager PIN",
			logger.Any("business_id", input.BusinessID),
			logger.Any("staff_id", input.StaffID),
			logger.Error(err),
		)
		return err
	}

	// A successful set clears the counter. The owner proved themselves, and
	// leaving a stale failure count behind would keep a legitimate next attempt
	// inside a backoff window or a lockout that no longer reflects any risk.
	s.resetPINFailures(ctx, store, input.BusinessID)
	return nil
}
