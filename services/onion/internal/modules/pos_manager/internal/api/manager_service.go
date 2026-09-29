package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

// invalidPINMessage is the single operator-facing text for a refused manager
// PIN. Every credential failure returns it, so the body cannot distinguish a
// wrong PIN from a locked-out business.
const invalidPINMessage = "Invalid PIN"

// setPINDeniedMessage is the single text for a refused manager PIN change. It
// is deliberately different from invalidPINMessage: the two endpoints refuse
// different things, and each one uses a constant, so neither reveals whether the
// business is locked out. Reusing the verify text here would be safe too, but it
// would tell an owner who mistyped their account password that their *manager*
// PIN was wrong.
const setPINDeniedMessage = "Could not verify your credentials"

// managerService is the application surface this handler needs. Declaring it
// here keeps the handler testable without a database, which is what lets the
// authorization and uniform-response tests assert status and body directly.
type managerService interface {
	VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.VerifyResult, error)
	SetPIN(ctx context.Context, input domain.SetPINInput) error
}

type ManagerAPI struct {
	Service managerService
}

func NewManagerAPI(service managerService) *ManagerAPI {
	return &ManagerAPI{Service: service}
}

// VerifyPIN checks a manager PIN for a POS manager override.
//
// The permission gate is checked before anything else, including before the
// tenant is resolved. Verifying a manager PIN is only meaningful as the second
// step of a POS action — a void, a refund, a drawer open — so the caller must
// already be able to take that action. Without this gate any authenticated user
// with access to the business could sit and guess manager PINs, and
// ResolveBusinessID is a tenant-membership check, not a POS entitlement.
func (a *ManagerAPI) VerifyPIN(ctx context.Context, req *posmanagerv1.VerifyPINRequest) (*posmanagerv1.VerifyPINResponse, error) {
	if !hasPOSManagerOverride(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	result, err := a.Service.VerifyPIN(ctx, businessID, strings.TrimSpace(req.GetPin()))
	if err != nil {
		if refusal := verifyPINRefusal(err); refusal != nil {
			return nil, refusal
		}
		logger.Ctx(ctx).Error("pos_manager verify-pin failed",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify PIN")
	}
	if result == nil || !result.Valid {
		return nil, onekiterrors.New(http.StatusUnauthorized, invalidPINMessage)
	}

	// staff_id is the manager the override is attributed to; that attribution is
	// the point of the endpoint and is only ever returned after a real
	// successful comparison.
	return &posmanagerv1.VerifyPINResponse{Valid: true, StaffId: result.StaffID, Message: "PIN verified"}, nil
}

// verifyPINRefusal maps every credential-shaped failure onto one response.
//
// VerifyPIN resolves an empty PIN, a business with no usable credential, a
// wrong PIN, an active lockout and an unreadable counter store to the single
// domain.ErrInvalidPIN sentinel. They must leave the client as one
// indistinguishable response: naming the cause would let a caller probe the
// lockout state of a business, and a lockout-specific status would let it poll
// that state directly. Returns nil for errors that are not credential failures,
// so the caller can surface them as server errors rather than pretending the PIN
// was wrong.
func verifyPINRefusal(verifyErr error) error {
	if errors.Is(verifyErr, domain.ErrInvalidPIN) {
		return onekiterrors.New(http.StatusUnauthorized, invalidPINMessage)
	}
	return nil
}

// hasPOSManagerOverride requires a POS-scoped caller who can take the action the
// override would authorise. It matches the guard the sibling pos_users verify
// route uses, so the two POS PIN surfaces cannot end up with different access.
func hasPOSManagerOverride(ctx context.Context) bool {
	device := authPkg.DeviceCtx(ctx)
	return device.User != nil && authPkg.HasAppPermission(
		device.User,
		authorization.AppPOS,
		"pos:orders:view",
	) && authPkg.HasAnyAppScope(device.User, authorization.AppPOS)
}

// SetPIN stores a manager's override PIN.
//
// The gate is the business owner and nothing less, and it is checked before the
// tenant is resolved for the same reason VerifyPIN's is.
//
// Using the VerifyPIN bar here would be the obvious mistake and a serious one.
// VerifyPIN only requires that the caller can take the POS action the override
// authorises, which every till operator clears; granting that same bar to the
// route that mints the PIN would let exactly the people the manager PIN exists
// to constrain provision it. A POS permission string is deliberately not used
// either: the permissions that administer employees and terminals are
// assignable to managers, and a manager who could mint a credential could then
// approve their own voids and refunds with it, which is the single outcome this
// control exists to prevent. The owner is the only identity whose authority to
// grant that is not itself derived from the thing being granted.
//
// Re-authentication with the owner's account password is the second half of the
// guard and lives in the service, so a stolen or left-open owner session cannot
// use a live token to mint a permanent override credential.
func (a *ManagerAPI) SetPIN(ctx context.Context, req *posmanagerv1.SetPINRequest) (*posmanagerv1.SetPINResponse, error) {
	if !canSetManagerPIN(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	setErr := a.Service.SetPIN(ctx, domain.SetPINInput{
		// The owner is the actor authorising the write, not the subject of the
		// credential. Conflating them would attribute a manager's PIN to the
		// owner in the audit record.
		OwnerID:         device.User.UserID,
		BusinessID:      businessID,
		StaffID:         req.GetStaffId(),
		PIN:             req.GetPin(),
		CurrentPassword: req.GetCurrentPassword(),
	})
	if setErr != nil {
		if refusal := setPINRefusal(setErr); refusal != nil {
			return nil, refusal
		}
		logger.Ctx(ctx).Error("pos_manager set-pin failed",
			logger.Any("business_id", businessID),
			logger.Any("staff_id", req.GetStaffId()),
			logger.Error(setErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to set PIN")
	}

	return &posmanagerv1.SetPINResponse{Success: true}, nil
}

// canSetManagerPIN authorises minting a manager override PIN: the business
// owner, with the POS app scoped so a business that has dropped POS cannot
// provision manager credentials through it.
func canSetManagerPIN(ctx context.Context) bool {
	device := authPkg.DeviceCtx(ctx)
	user := device.User
	if user == nil || user.EffectiveAccess == nil || !user.EffectiveAccess.IsOwner {
		return false
	}
	return authPkg.HasAnyAppScope(user, authorization.AppPOS)
}

// setPINRefusal maps the set path's failures onto responses.
//
// ErrSetPINNotAuthorized covers a wrong owner password, an active backoff and a
// locked-out business, and all three answer with one 401 and one body: this
// route is a password check, so naming which of the three occurred would let a
// caller probe the lockout state and would turn it into a password-guessing
// oracle. The policy and shape failures are the caller's own input, so they are
// named rather than hidden — they are answered before any credential is read.
//
// Returns nil for errors that are not refusals so the caller can surface them
// as server errors rather than disguising an outage as a rejected PIN.
func setPINRefusal(setErr error) error {
	switch {
	case errors.Is(setErr, domain.ErrSetPINNotAuthorized):
		return onekiterrors.New(http.StatusUnauthorized, setPINDeniedMessage)
	case errors.Is(setErr, domain.ErrWeakPIN):
		return onekiterrors.New(http.StatusBadRequest, setErr.Error())
	case errors.Is(setErr, domain.ErrInvalidSetPIN):
		return onekiterrors.New(http.StatusBadRequest, "Invalid manager PIN request")
	default:
		return nil
	}
}
