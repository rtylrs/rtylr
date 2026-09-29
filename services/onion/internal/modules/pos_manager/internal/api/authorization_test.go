package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

// recordingService captures whether the credential path was reached at all.
type recordingService struct {
	calls    int
	setCalls int
	setInput domain.SetPINInput
	result   *domain.VerifyResult
	err      error
	setErr   error
}

func (s *recordingService) VerifyPIN(context.Context, int64, string) (*domain.VerifyResult, error) {
	s.calls++
	return s.result, s.err
}

func (s *recordingService) SetPIN(_ context.Context, input domain.SetPINInput) error {
	s.setCalls++
	s.setInput = input
	return s.setErr
}

func managerStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

// posCtx builds a context for a caller that is logged in and holds exactly the
// given POS permissions. The permissions belong on EffectiveAccess.Apps[]
// .Permissions — that is what authPkg.HasAppPermission reads.
func posCtx(permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppPOS,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	})
}

// skipInLegacyAuthorizationMode documents that AUTHORIZATION_V2_MODE=legacy
// makes HasAppPermission accept any licensed POS app regardless of the
// permission list, so the "denied without the permission" cases cannot be
// asserted in that mode. The same caveat applies to the sibling pos_users gate.
func skipInLegacyAuthorizationMode(t *testing.T) {
	t.Helper()
	if authPkg.LegacyAuthorizationMode() {
		t.Skip("AUTHORIZATION_V2_MODE=legacy grants POS app access without checking permissions")
	}
}

// A caller who is logged in but cannot take the POS action the override would
// authorise must be refused before the PIN is ever checked. Pre-fix the only
// guards were "is logged in" and a tenant-membership lookup, so any
// authenticated user of the business could brute-force manager PINs.
func TestVerifyPIN_RequiresPOSOrderViewPermission(t *testing.T) {
	skipInLegacyAuthorizationMode(t)

	cases := map[string][]string{
		"no permissions at all":      {},
		"unrelated permission":       {"pos:employees:manage"},
		"orders create but not view": {"pos:orders:create"},
	}
	for name, perms := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &recordingService{result: &domain.VerifyResult{Valid: true, StaffID: 1}}
			api := &ManagerAPI{Service: svc}

			_, err := api.VerifyPIN(posCtx(perms...), &posmanagerv1.VerifyPINRequest{Pin: "123456"})
			if managerStatus(err) != http.StatusForbidden {
				t.Fatalf("returned %v, want 403 before the PIN is checked", err)
			}
			if svc.calls != 0 {
				t.Errorf("the credential path was reached %d times by an unauthorised caller", svc.calls)
			}
		})
	}
}

// The gate is asserted directly rather than through the handler: a permitted
// caller gets past it and then fails at tenant resolution, which needs a
// database and so is not observable in a unit test.
func TestHasPOSManagerOverride_PermittedCallerPassesTheGate(t *testing.T) {
	skipInLegacyAuthorizationMode(t)
	if !hasPOSManagerOverride(posCtx("pos:orders:view")) {
		t.Error("a caller holding pos:orders:view was refused by the gate")
	}
}

func TestHasPOSManagerOverride_DeniesWithoutThePermission(t *testing.T) {
	skipInLegacyAuthorizationMode(t)
	if hasPOSManagerOverride(posCtx()) {
		t.Error("a caller with no POS permissions passed the gate")
	}
	if hasPOSManagerOverride(posCtx("pos:employees:manage")) {
		t.Error("a caller holding an unrelated POS permission passed the gate")
	}
}

func TestHasPOSManagerOverride_DeniesAnonymous(t *testing.T) {
	if hasPOSManagerOverride(context.Background()) {
		t.Error("an anonymous caller passed the gate")
	}
}

// The gate must run before the tenant resolver, not after it. A denied caller
// gets 403 with no database reachable at all, which only holds if the
// permission check short-circuits first.
func TestVerifyPIN_PermissionGateRunsBeforeTenantResolution(t *testing.T) {
	skipInLegacyAuthorizationMode(t)

	svc := &recordingService{result: &domain.VerifyResult{Valid: true, StaffID: 1}}
	api := &ManagerAPI{Service: svc}

	// No database is available here, so tenant.ResolveBusinessID would fail.
	// Reaching 403 with the credential path untouched shows the gate answered
	// first.
	_, err := api.VerifyPIN(posCtx("pos:employees:manage"), &posmanagerv1.VerifyPINRequest{Pin: "123456"})
	if managerStatus(err) != http.StatusForbidden {
		t.Fatalf("returned %v, want 403", err)
	}
	if svc.calls != 0 {
		t.Errorf("the credential path ran %d times for a denied caller", svc.calls)
	}
}

// An anonymous caller must be refused.
func TestVerifyPIN_AnonymousCallerRefused(t *testing.T) {
	svc := &recordingService{result: &domain.VerifyResult{Valid: true, StaffID: 1}}
	api := &ManagerAPI{Service: svc}

	_, err := api.VerifyPIN(context.Background(), &posmanagerv1.VerifyPINRequest{Pin: "123456"})
	if managerStatus(err) != http.StatusForbidden {
		t.Fatalf("anonymous caller returned %v, want 403", err)
	}
	if svc.calls != 0 {
		t.Error("an anonymous caller reached the credential path")
	}
}

// Every credential-shaped failure must leave the client as one indistinguishable
// response. Naming the cause would let a caller probe a business's lockout
// state, and a lockout-specific status would let it poll that state directly.
func TestVerifyPINRefusal_UniformAcrossEveryCredentialFailure(t *testing.T) {
	baseline := verifyPINRefusal(domain.ErrInvalidPIN)
	if baseline == nil {
		t.Fatal("ErrInvalidPIN must produce a refusal")
	}
	wantStatus := managerStatus(baseline)
	var wantErr *onekithttp.Error
	if !errors.As(baseline, &wantErr) {
		t.Fatalf("baseline returned %T, want *onekiterrors.Error", baseline)
	}

	// 429 and 423 both name the lockout, which is the specific outcome this
	// mapping must never reveal.
	for _, status := range []int{http.StatusTooManyRequests, http.StatusLocked} {
		if wantStatus == status {
			t.Errorf("verify answered %d; a lockout-specific status confirms the state to the caller", status)
		}
	}

	for _, cause := range []error{
		domain.ErrInvalidPIN,
		fmt.Errorf("repository: %w", domain.ErrInvalidPIN),
		fmt.Errorf("locked: %w", domain.ErrInvalidPIN),
	} {
		got := verifyPINRefusal(cause)
		if got == nil {
			t.Errorf("%v produced no refusal; a credential failure must not fall through to a 500", cause)
			continue
		}
		if status := managerStatus(got); status != wantStatus {
			t.Errorf("%v: status = %d, want %d", cause, status, wantStatus)
		}
		var gotErr *onekithttp.Error
		if !errors.As(got, &gotErr) {
			t.Errorf("%v: returned %T, want *onekiterrors.Error", cause, got)
			continue
		}
		if gotErr.GetMessage() != wantErr.GetMessage() {
			t.Errorf("%v: body = %q, want %q", cause, gotErr.GetMessage(), wantErr.GetMessage())
		}
	}
}

// A transport or database failure must not be laundered into a credential
// refusal: that would make an outage look like a wrong PIN to the operator and
// to monitoring.
func TestVerifyPINRefusal_RealErrorsAreNotCredentialFailures(t *testing.T) {
	if refusal := verifyPINRefusal(errors.New("connection refused")); refusal != nil {
		t.Errorf("a transport failure produced %v; only ErrInvalidPIN is a credential refusal", refusal)
	}
}

// ownerCtx builds a context for the business owner. Owner is resolved on
// EffectiveAccess.IsOwner in both the v2 and the legacy resolver, so the set-PIN
// gate is asserted the same way in either authorization mode.
func ownerCtx(withPOS bool) context.Context {
	apps := []authorization.EffectiveAppAccess{}
	if withPOS {
		apps = append(apps, authorization.EffectiveAppAccess{
			AppID:       authorization.AppPOS,
			Licensed:    true,
			HasAccess:   true,
			Permissions: []string{"pos:orders:view"},
		})
	}
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     9,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsOwner: true,
			Apps:    apps,
		},
	})
}

// Minting a manager PIN is the strongest thing this module does, and it must not
// be reachable by anyone who is not the business owner. The pre-fix module had no
// write path at all; adding one behind the same bar as VerifyPIN would have
// handed every till operator the ability to provision the credential that
// constrains them.
func TestCanSetManagerPIN_OnlyTheOwnerPasses(t *testing.T) {
	if !canSetManagerPIN(ownerCtx(true)) {
		t.Error("the business owner was refused the ability to set a manager PIN")
	}

	// Every non-owner shape: no user at all, a user with no resolved access, a
	// manager who administers employees, and a plain cashier.
	denied := map[string]context.Context{
		"anonymous": context.Background(),
		"no effective access": authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
			BusinessID: 9, UserID: 7,
		}),
		"pos manager with employees:manage": posCtx("pos:employees:manage"),
		"cashier":                           posCtx("pos:orders:view"),
	}
	for name, ctx := range denied {
		t.Run(name, func(t *testing.T) {
			if canSetManagerPIN(ctx) {
				t.Error("a non-owner passed the manager PIN gate")
			}
		})
	}
}

// The VerifyPIN bar must not transfer to the set path. A caller holding exactly
// what VerifyPIN requires is still refused, which is the regression this route
// would introduce if the two gates were ever unified.
func TestCanSetManagerPIN_VerifyPermissionAloneIsNotEnough(t *testing.T) {
	if canSetManagerPIN(posCtx("pos:orders:view")) {
		t.Error("pos:orders:view alone granted the ability to mint a manager PIN")
	}
}

// An owner whose business has dropped POS must not be able to provision manager
// credentials through a POS route.
func TestCanSetManagerPIN_OwnerWithoutPOSScopeIsRefused(t *testing.T) {
	if canSetManagerPIN(ownerCtx(false)) {
		t.Error("an owner with no POS scope passed the manager PIN gate")
	}
}

// The gate runs before tenant resolution, so a denied caller gets 403 with no
// database reachable and the set path never runs.
func TestSetPIN_PermissionGateRunsBeforeTenantResolution(t *testing.T) {
	svc := &recordingService{}
	api := &ManagerAPI{Service: svc}

	// No database is reachable here, so tenant.ResolveBusinessID would fail.
	// Reaching 403 with the set path untouched shows the gate answered first.
	_, err := api.SetPIN(posCtx("pos:orders:view"), &posmanagerv1.SetPINRequest{
		StaffId: 1, Pin: "123456", CurrentPassword: "hunter2000",
	})
	if managerStatus(err) != http.StatusForbidden {
		t.Fatalf("returned %v, want 403", err)
	}
	if svc.setCalls != 0 {
		t.Errorf("the set path ran %d times for a denied caller", svc.setCalls)
	}
}

func TestSetPIN_AnonymousCallerRefused(t *testing.T) {
	svc := &recordingService{}
	api := &ManagerAPI{Service: svc}

	_, err := api.SetPIN(context.Background(), &posmanagerv1.SetPINRequest{
		StaffId: 1, Pin: "123456", CurrentPassword: "hunter2000",
	})
	if managerStatus(err) != http.StatusForbidden {
		t.Fatalf("anonymous caller returned %v, want 403", err)
	}
	if svc.setCalls != 0 {
		t.Error("an anonymous caller reached the set path")
	}
}

// The set path is a password check, so a wrong owner password, an active backoff
// and a locked-out business must leave the client as one indistinguishable 401.
// Splitting them would both reveal the lockout state and turn this route into a
// password-guessing oracle.
func TestSetPINRefusal_UniformAcrossEveryCredentialFailure(t *testing.T) {
	baseline := setPINRefusal(domain.ErrSetPINNotAuthorized)
	if baseline == nil {
		t.Fatal("ErrSetPINNotAuthorized must produce a refusal")
	}
	wantStatus := managerStatus(baseline)
	var wantErr *onekithttp.Error
	if !errors.As(baseline, &wantErr) {
		t.Fatalf("baseline returned %T, want *onekiterrors.Error", baseline)
	}

	for _, status := range []int{http.StatusTooManyRequests, http.StatusLocked} {
		if wantStatus == status {
			t.Errorf("set-pin answered %d; a lockout-specific status confirms the state to the caller", status)
		}
	}

	for _, cause := range []error{
		domain.ErrSetPINNotAuthorized,
		fmt.Errorf("locked: %w", domain.ErrSetPINNotAuthorized),
		fmt.Errorf("owner credential: %w", domain.ErrSetPINNotAuthorized),
	} {
		got := setPINRefusal(cause)
		if got == nil {
			t.Errorf("%v produced no refusal; a credential failure must not fall through to a 500", cause)
			continue
		}
		if status := managerStatus(got); status != wantStatus {
			t.Errorf("%v: status = %d, want %d", cause, status, wantStatus)
		}
		var gotErr *onekithttp.Error
		if !errors.As(got, &gotErr) {
			continue
		}
		if gotErr.GetMessage() != wantErr.GetMessage() {
			t.Errorf("%v: body = %q, want %q", cause, gotErr.GetMessage(), wantErr.GetMessage())
		}
	}
}

// A weak PIN is the caller's own input, so it is named rather than hidden. The
// operator needs to know which rule they broke.
func TestSetPINRefusal_WeakPINIsNamed(t *testing.T) {
	got := setPINRefusal(domain.ErrWeakPIN)
	if status := managerStatus(got); status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 so the owner is told their PIN is too weak", status)
	}
}

// An outage must not be laundered into a rejected PIN: that would hide a
// database failure from the operator and from monitoring.
func TestSetPINRefusal_RealErrorsAreNotCredentialFailures(t *testing.T) {
	if refusal := setPINRefusal(errors.New("connection refused")); refusal != nil {
		t.Errorf("a transport failure produced %v; it must surface as a server error", refusal)
	}
	if refusal := setPINRefusal(domain.ErrSetPINStoreUnavailable); refusal != nil {
		t.Errorf("an unreadable credential store produced %v; it must surface as a server error", refusal)
	}
}
