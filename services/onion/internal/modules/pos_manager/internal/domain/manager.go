package domain

import "errors"

var ErrInvalidPIN = errors.New("invalid manager PIN")

// MaxManagerPINCandidates bounds how many stored manager PIN hashes a single
// verification may compare. The verify request carries only a PIN and no
// manager identity, so the credential cannot be found by index seek against a
// slow KDF; every candidate at the business is compared instead, at roughly
// 210ms per bcrypt-cost-12 hash.
//
// The aggregate cost is already bounded by the per-business lockout (5 attempts
// per 15 minutes), so this cap is not there to stop CPU amplification. It is
// there to keep one request's latency finite. A business with more candidates
// than this fails closed rather than being silently truncated, because
// truncating would let an attacker park a decoy row above the cap and have the
// real manager's PIN skipped.
//
// TODO: a peppered blind index (HMAC of the PIN under a server secret held
// outside the database) would restore an O(1) lookup and let the cap go away.
// It needs a configuration secret this module does not own.
const MaxManagerPINCandidates = 32

type VerifyResult struct {
	Valid      bool   `json:"valid"`
	StaffID    int64  `json:"staff_id"`
	OwnerID    int64  `json:"owner_id"`
	BusinessID int64  `json:"business_id"`
	Name       string `json:"name"`
}

// ManagerPINCandidate is one stored manager credential, as read for
// verification. The hash is compared with bcrypt; the plaintext is never
// stored or returned.
type ManagerPINCandidate struct {
	OwnerID int64
	StaffID int64
	PinHash string
}

// SetPINInput is a request to set or rotate a manager's override PIN.
//
// OwnerID is the business owner who authorised the write, not the manager the
// credential belongs to: the two are different identities and conflating them
// would let a manager attribute their own credential to the owner.
type SetPINInput struct {
	OwnerID    int64
	BusinessID int64
	StaffID    int64
	PIN        string
	// CurrentPassword is the owner's account password, required so a stolen or
	// left-open owner session cannot mint a manager PIN.
	CurrentPassword string
}

var (
	// ErrInvalidSetPIN reports a malformed set request. It is deliberately
	// distinct from ErrInvalidPIN, which is the uniform verify refusal: naming
	// this one tells the operator which request was wrong, and it leaks nothing
	// because it is answered before any credential is touched.
	ErrInvalidSetPIN = errors.New("invalid manager PIN request")
	// ErrSetPINNotAuthorized reports that the caller's account password did not
	// match. It is a credential failure, so it is answered identically to
	// ErrInvalidPIN and counts against the same per-business lockout.
	ErrSetPINNotAuthorized = errors.New("manager PIN change not authorized")
	// ErrSetPINStoreUnavailable reports that the owner's credential could not be
	// read. An unknown state is never a satisfied check, so this denies.
	ErrSetPINStoreUnavailable = errors.New("manager PIN authorization state unavailable")
)
