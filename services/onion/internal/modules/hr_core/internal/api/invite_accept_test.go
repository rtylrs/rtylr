package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	inviteBusinessID  int64 = 7001
	otherBusinessID   int64 = 7002
	inviteeEmployeeID int64 = 8001
	otherEmployeeID   int64 = 8002
	inviteeUserID     int64 = 8101
	attackerUserID    int64 = 8102
	otherUserID       int64 = 8103
	linkedEmployeeID  int64 = 8003
	inviteeEmail            = "invitee.8101@example.invalid"
	attackerEmail           = "attacker.8102@example.invalid"
	otherInviteeEmail       = "other.invitee.8103@example.invalid"
)

func inviteCaller(userID int64) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: userID, BusinessID: inviteBusinessID})
}

func decodeAcceptRequest(t *testing.T, body string) *AcceptInviteRequest {
	t.Helper()
	req := &AcceptInviteRequest{}
	if err := json.Unmarshal([]byte(body), req); err != nil {
		t.Fatalf("decode accept body %s: %v", body, err)
	}
	return req
}

// The acceptance used to read the account to link straight out of the request
// body, on a route registered as anonymous, so anyone holding a token could
// name the account that becomes the employee of a business they were never
// invited to. There is no caller to identify in an anonymous request, so it is
// refused before any work happens.
func TestAcceptInviteHandler_AnonymousRequestIsRejected(t *testing.T) {
	handler := &AcceptInviteHandler{}

	resp, err := handler.Serve(context.Background(), &AcceptInviteRequest{Token: "any-token"})
	if resp.Success {
		t.Fatal("an anonymous acceptance reported success")
	}
	if status := oneKitStatus(err); status != http.StatusForbidden && status != http.StatusUnauthorized {
		t.Fatalf("anonymous acceptance status = %d, want 401 or 403 (err=%v)", status, err)
	}
}

// A user_id in the body must not survive as a field the handler can read: the
// account linked is the authenticated caller's even when the body names another.
func TestAcceptInviteHandler_BodyCannotSelectTheAccount(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addAccount(attackerUserID, attackerEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")

	req := decodeAcceptRequest(t, `{"token":"token-invitee","user_id":8102}`)
	handler := &AcceptInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	resp, err := handler.Serve(inviteCaller(inviteeUserID), req)
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected the addressed invitee to be able to accept")
	}

	employee, _ := store.employee(inviteeEmployeeID)
	if employee.userID != inviteeUserID {
		t.Fatalf("linked account = %d, want the authenticated caller %d; a body user_id must never choose the account", employee.userID, inviteeUserID)
	}
}

// The invite names the email it was sent to. A different authenticated account
// - the inviter, whoever intercepted the token - must not be able to claim it,
// and must not burn it for the person it was addressed to.
func TestAcceptInviteHandler_InviteAddressedToAnotherAccountIsRejected(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addAccount(attackerUserID, attackerEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")

	handler := &AcceptInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	resp, err := handler.Serve(inviteCaller(attackerUserID), &AcceptInviteRequest{Token: "token-invitee"})
	if resp.Success {
		t.Fatal("an account the invite was not addressed to reported success")
	}
	if status := oneKitStatus(err); status != http.StatusForbidden && status != http.StatusUnauthorized {
		t.Fatalf("wrong-recipient status = %d, want 401 or 403 (err=%v)", status, err)
	}

	employee, _ := store.employee(inviteeEmployeeID)
	if employee.userID != 0 || employee.status != "invited" {
		t.Fatalf("employee = %+v, want an unlinked invited row after a refused acceptance", employee)
	}
	if store.tokenConsumed("token-invitee") {
		t.Fatal("a refused acceptance consumed the invite, denying it to its recipient")
	}

	resp, err = handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"})
	if err != nil || !resp.Success {
		t.Fatalf("the addressed invitee could not accept the preserved token: resp=%+v err=%v", resp, err)
	}
}

// An invite is a single-use credential: the second acceptance must not re-link
// the row, which is how a leaked token would stay replayable.
func TestAcceptInviteHandler_TokenIsSingleUse(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")

	handler := &AcceptInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	if resp, err := handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"}); err != nil || !resp.Success {
		t.Fatalf("first acceptance: resp=%+v err=%v", resp, err)
	}

	resp, err := handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"})
	if resp.Success {
		t.Fatal("a consumed invite was accepted a second time")
	}
	if status := oneKitStatus(err); status != http.StatusBadRequest && status != http.StatusForbidden {
		t.Fatalf("replayed token status = %d, want 400 or 403 (err=%v)", status, err)
	}
}

// A token selects the tenant it was minted for, and only that tenant: accepting
// a business-B invite must not write anything into business A.
func TestAcceptInviteHandler_AcceptanceStaysInsideTheInvitedBusiness(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addAccount(otherUserID, otherInviteeEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")
	store.addInvited(otherBusinessID, otherEmployeeID, otherInviteeEmail, "token-other")

	handler := &AcceptInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	if resp, err := handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"}); err != nil || !resp.Success {
		t.Fatalf("acceptance: resp=%+v err=%v", resp, err)
	}

	other, _ := store.employee(otherEmployeeID)
	if other.userID != 0 || other.status != "invited" {
		t.Fatalf("employee in the other business = %+v, want an untouched invited row", other)
	}
	if store.tokenConsumed("token-other") {
		t.Fatal("accepting one business's invite consumed another business's invite")
	}

	// A token from another business links only into that business: the first
	// tenant's already-active row must not be re-linked or re-consumed.
	if resp, err := handler.Serve(inviteCaller(otherUserID), &AcceptInviteRequest{Token: "token-other"}); err != nil || !resp.Success {
		t.Fatalf("acceptance of the other business's invite: resp=%+v err=%v", resp, err)
	}
	invited, _ := store.employee(inviteeEmployeeID)
	if invited.userID != inviteeUserID || invited.status != "active" {
		t.Fatalf("employee in the invited business = %+v, want the first acceptance to stand", invited)
	}
	if got := store.tokenConsumptions("token-invitee"); got != 1 {
		t.Fatalf("token-invitee consumed %d times, want exactly 1", got)
	}
}

// A link that cannot succeed - the account is already on another employee row -
// must leave the invite usable, exactly as it did for the wrong employee case.
func TestAcceptInviteHandler_FailedLinkLeavesTokenUsable(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")
	store.mu.Lock()
	store.employees[linkedEmployeeID] = &inviteStoreEmployee{
		businessID: inviteBusinessID, email: "linked.8101@example.invalid", status: "active", userID: inviteeUserID,
	}
	store.mu.Unlock()

	handler := &AcceptInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	resp, err := handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"})
	if resp.Success {
		t.Fatal("an acceptance onto an already-linked account reported success")
	}
	if status := oneKitStatus(err); status != http.StatusConflict && status != http.StatusBadRequest {
		t.Fatalf("already-linked status = %d, want 409 or 400 (err=%v)", status, err)
	}
	if store.tokenConsumed("token-invitee") {
		t.Fatal("a failed link consumed the invite")
	}

	store.unlinkAccount(inviteeUserID)
	if resp, err := handler.Serve(inviteCaller(inviteeUserID), &AcceptInviteRequest{Token: "token-invitee"}); err != nil || !resp.Success {
		t.Fatalf("retry with the preserved token: resp=%+v err=%v", resp, err)
	}
}

// Verify is registered next to accept and used to answer with the employee and
// business a token points at, to whoever presented it.
func TestVerifyInviteHandler_AnonymousRequestIsRejected(t *testing.T) {
	handler := &VerifyInviteHandler{}

	if _, err := handler.Serve(context.Background(), &VerifyInviteRequest{Token: "any-token"}); oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("anonymous verify status = %d, want 403 (err=%v)", oneKitStatus(err), err)
	}
}

func TestVerifyInviteHandler_DoesNotDiscloseTenantIdentifiers(t *testing.T) {
	store := newInviteStore()
	store.addAccount(inviteeUserID, inviteeEmail)
	store.addInvited(inviteBusinessID, inviteeEmployeeID, inviteeEmail, "token-invitee")

	handler := &VerifyInviteHandler{Service: application.NewService(context.Background(), openInviteStore(t, store))}

	resp, err := handler.Serve(inviteCaller(attackerUserID), &VerifyInviteRequest{Token: "token-invitee"})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if !resp.Valid {
		t.Fatal("expected a valid invite to verify as valid")
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal verify response: %v", err)
	}
	for _, leaked := range []string{"employee_id", "owner_id", "business_id"} {
		if strings.Contains(string(body), leaked) {
			t.Fatalf("verify response %s exposes %s", body, leaked)
		}
	}
}
