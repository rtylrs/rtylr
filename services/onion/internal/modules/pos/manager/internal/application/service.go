package application

import (
	"context"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	// verifier is the narrowed repository view used by the credential path and
	// attempts the durable brute-force counter. They are separate fields so the
	// verify path can be exercised without a database.
	verifier managerPINVerifier
	// writer and credentials are the write half of the module: the credential
	// store and the owner's account credential used to re-authorise a change.
	writer      managerPINWriter
	credentials ownerCredentialVerifier
	attempts    PinAttemptStore
	// now is the clock used to decide whether a business is inside its backoff
	// window. Overridable in tests so the schedule can be asserted without
	// sleeping.
	now func() time.Time
}

func NewService(verifier managerPINVerifier, writer managerPINWriter, credentials ownerCredentialVerifier, attempts PinAttemptStore) *Service {
	return &Service{
		verifier:    verifier,
		writer:      writer,
		credentials: credentials,
		attempts:    attempts,
		now:         time.Now,
	}
}

func (s *Service) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// VerifyPIN checks a manager PIN and, on success, returns the manager it belongs
// to. That manager identity is what the override is attributed to, so this is
// the endpoint worth brute-forcing, and the guard is deliberately blunt: a
// durable per-business attempt counter with exponential backoff, and a single
// uniform failure for every reason the check can fail.
//
// The uniform failure is load-bearing, not cosmetic. An empty candidate set, a
// wrong PIN, a business with no usable credential, an active lockout and an
// unreadable counter store all return domain.ErrInvalidPIN, and all of them cost
// a bcrypt comparison, so status, body and latency are indistinguishable.
//
// One leak is accepted and deliberate: a business with no candidate rows costs
// one dummy comparison while a business with N rows costs N. That reveals how
// many manager credentials a business holds, which is a property of the
// caller's own tenant and not attacker-controllable — a materially weaker signal
// than "does this specific user hold a PIN", which is the one the dummy compare
// exists to close.
func (s *Service) VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.VerifyResult, error) {
	pin = strings.TrimSpace(pin)
	if businessID <= 0 || pin == "" {
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}

	store := s.attempts

	// Fail closed when the counter cannot be read. A counter we cannot read is a
	// counter we cannot enforce, and skipping the check would restore the
	// unbounded guess budget this guard exists to remove. Locked-out callers land
	// here too, so a correct PIN is refused until the window expires.
	if store == nil {
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}
	retryAt, retryErr := store.PinRetryAt(ctx, businessID)
	if retryErr != nil {
		logger.Ctx(ctx).Error("manager pin attempt counter unavailable; denying verification",
			logger.Any("business_id", businessID),
			logger.Error(retryErr),
		)
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}
	if !retryAt.IsZero() && s.clock().Before(retryAt) {
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}

	verifier := s.verifierOrDefault()
	candidates, err := verifier.ListManagerPINHashes(ctx, businessID)
	if err != nil {
		// A read failure is an infrastructure fault, not a credential verdict.
		// Reporting it as one would make an outage look like a wrong PIN to the
		// operator and to monitoring.
		return nil, err
	}

	if len(candidates) == 0 {
		s.recordPINFailure(ctx, store, businessID)
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}
	if len(candidates) > domain.MaxManagerPINCandidates {
		// Fail closed rather than truncate: a truncated scan would let an
		// attacker park a decoy row above the cap and have the real manager's
		// PIN silently skipped.
		logger.Ctx(ctx).Error("manager PIN candidate count exceeds the verification budget; denying",
			logger.Any("business_id", businessID),
			logger.Any("candidates", len(candidates)),
			logger.Any("max", domain.MaxManagerPINCandidates),
		)
		burnManagerPINCompare(pin)
		return nil, domain.ErrInvalidPIN
	}

	for i, candidate := range candidates {
		if checkManagerPINHash(pin, candidate.PinHash) {
			// Burn out the rest of the scan so a match at candidate 0 costs the
			// same as a match at the last candidate.
			for j := i + 1; j < len(candidates); j++ {
				burnManagerPINCompare(pin)
			}
			s.resetPINFailures(ctx, store, businessID)
			return &domain.VerifyResult{
				Valid:      true,
				StaffID:    candidate.StaffID,
				OwnerID:    candidate.OwnerID,
				BusinessID: businessID,
			}, nil
		}
	}

	// Every candidate was compared, so this path already cost a full scan.
	s.recordPINFailure(ctx, store, businessID)
	return nil, domain.ErrInvalidPIN
}

func (s *Service) verifierOrDefault() managerPINVerifier {
	if s.verifier != nil {
		return s.verifier
	}
	return missingPINVerifier{}
}

// recordPINFailure increments the brute-force counter, logging rather than
// returning a counter-store failure: the PIN is already being rejected, so
// there is no safer outcome to switch to and no oracle to avoid.
func (s *Service) recordPINFailure(ctx context.Context, store PinAttemptStore, businessID int64) {
	if store == nil {
		return
	}
	if _, err := store.RecordPINFailure(ctx, businessID); err != nil {
		logger.Ctx(ctx).Error("failed to record manager pin verification failure",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
	}
}

// resetPINFailures clears the counter after a real success, so a manager who
// eventually types their PIN correctly is not left permanently locked out.
func (s *Service) resetPINFailures(ctx context.Context, store PinAttemptStore, businessID int64) {
	if store == nil {
		return
	}
	if err := store.ResetPINFailures(ctx, businessID); err != nil {
		logger.Ctx(ctx).Warn("failed to clear manager pin attempt counter after success",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
	}
}

// missingPINVerifier stands in for a Service built without a repository, as the
// zero-value unit tests do. The read fails, which the verify path surfaces as a
// server error rather than a credential verdict.
type missingPINVerifier struct{}

func (missingPINVerifier) ListManagerPINHashes(context.Context, int64) ([]domain.ManagerPINCandidate, error) {
	return nil, domain.ErrNoManagerPINs
}

type credentialVerifierFunc func(context.Context, int64, string) (bool, error)

func (f credentialVerifierFunc) VerifyAccountPassword(ctx context.Context, id int64, password string) (bool, error) {
	return f(ctx, id, password)
}

func (s *Service) WithPasswordVerifier(read func(context.Context, int64, string) (bool, error)) {
	if read != nil {
		s.credentials = credentialVerifierFunc(read)
	}
}
