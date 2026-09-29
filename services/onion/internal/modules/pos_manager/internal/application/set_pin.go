package application

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

// managerPINWriter is the slice of the repository SetPIN needs: the credential
// write. Narrowing it keeps the set path testable without a database and stops
// it growing new reads by accident.
type managerPINWriter interface {
	SetManagerPIN(ctx context.Context, input domain.SetPINInput, pinHash string, events ...audit.Event) error
}

// ownerCredentialReader reads the account credential the owner must re-enter.
type ownerCredentialReader interface {
	GetUserPasswordHash(ctx context.Context, userID int64) (string, error)
}

// SetPIN stores a manager's override PIN, replacing any credential they had.
//
// The authority to do this sits far above the authority to *use* the PIN.
// VerifyPIN only requires that the caller can take the POS action the override
// would authorise, which is a low bar any till operator clears; minting a
// credential that approves voids and refunds is a different and much stronger
// power, and it is therefore re-authorised here by the owner's account password.
// Without that second factor a stolen or left-open owner session could hand an
// attacker the manager override permanently.
//
// The caller's own account password is required rather than a current manager
// PIN. A manager PIN would be circular here: the migration that made these
// credentials safe invalidated every existing one, so requiring one to mint the
// first would make the route unreachable exactly when it is needed, and it would
// also let anyone who already holds a PIN mint more of them, which is the wrong
// direction for a root-of-trust credential.
//
// Every refusal that could be reached by guessing costs a bcrypt comparison and
// returns one indistinguishable 401, and a wrong owner password is counted
// against the same durable per-business lockout VerifyPIN uses. That endpoint is
// otherwise an unthrottled password oracle on an authenticated route: a session
// thief who could reach it would otherwise get unlimited guesses at the owner's
// password.
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

	hash, err := credentials.GetUserPasswordHash(ctx, input.OwnerID)
	if err != nil {
		// A credential read failure is an infrastructure fault, not a verdict on
		// the caller. Reporting it as one would make an outage look like a wrong
		// password to the operator and to monitoring.
		logger.Ctx(ctx).Error("failed to read owner credential for manager PIN change",
			logger.Any("business_id", input.BusinessID),
			logger.Error(err),
		)
		return domain.ErrSetPINStoreUnavailable
	}
	if !strings.HasPrefix(hash, bcryptHashPrefix) {
		// Nothing usable is stored, so there is no real comparison to make. Spend
		// a dummy one so this path costs the same KDF time as a genuine check and
		// cannot be told apart from it by latency.
		passwordutils.CheckDummyHash(input.CurrentPassword)
		s.recordPINFailure(ctx, store, input.BusinessID)
		return domain.ErrSetPINNotAuthorized
	}
	if !passwordutils.CheckPasswordHash(input.CurrentPassword, hash) {
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

// bcryptHashPrefix is the version marker every bcrypt hash starts with. It is
// used only to decide whether a stored credential is worth comparing, because
// bcrypt rejects a malformed hash immediately and cheaply.
const bcryptHashPrefix = "$2"
