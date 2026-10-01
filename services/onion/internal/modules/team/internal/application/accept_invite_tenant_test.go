package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"
)

// These exercise AcceptInvite against a scripted database/sql driver rather than
// MySQL, so the tenant-binding guarantee is checked on every `go test` run and
// the refusal path can be asserted to have written nothing at all.

const (
	inviteOwnerID    int64 = 7001
	inviteeID        int64 = 7002
	inviteBusinessID int64 = 7003
	otherBusinessID  int64 = 7004
)

func inviteRowScript(inviteBusiness any) scriptedQuery {
	return scriptedQuery{
		match:   containsAny("FROM invite_tokens"),
		columns: []string{"user_id", "owner_id", "business_id", "expires_at", "used_at"},
		rows: [][]driver.Value{{
			inviteeID, inviteOwnerID, inviteBusiness, time.Now().Add(time.Hour), nil,
		}},
	}
}

// The owner's business lookup answers "this owner owns no business".
var noBusinessForOwner = scriptedQuery{
	match:   containsAny("SELECT id FROM businesses WHERE owner_id"),
	columns: []string{"id"},
	err:     sql.ErrNoRows,
}

func businessInfoScript(ownerID int64) scriptedQuery {
	return scriptedQuery{
		match:   containsAny("SELECT name, owner_id FROM businesses WHERE id = ?"),
		columns: []string{"name", "owner_id"},
		rows:    [][]driver.Value{{"Invited Business", ownerID}},
	}
}

func statementSummary(statements []recordedStatement) string {
	parts := make([]string, 0, len(statements))
	for _, statement := range statements {
		parts = append(parts, statement.query)
	}
	return strings.Join(parts, "\n")
}

func argAt(args []driver.Value, i int) driver.Value {
	if i >= len(args) {
		return nil
	}
	return args[i]
}

// The owner's business lookup answers "this owner owns some other business" —
// the case a `LIMIT 1` lookup gets wrong for an owner with more than one, and
// the case the old code substituted users.business_id into.
var businessForOwner = scriptedQuery{
	match:   containsAny("SELECT id FROM businesses WHERE owner_id"),
	columns: []string{"id"},
	rows:    [][]driver.Value{{otherBusinessID}},
}

// A member who already belongs to one business is invited to another. The
// accept uses the invitation's business and verifies it belongs to its issuer.
func TestAcceptInviteWritesUnderTheInviteBusiness(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		businessInfoScript(inviteOwnerID),
	})
	service := NewService(context.Background(), db)

	userID, err := service.AcceptInvite(context.Background(), "live-token", acceptInvitePassword)
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if userID != inviteeID {
		t.Fatalf("AcceptInvite returned user %d, want %d", userID, inviteeID)
	}

	var provisioningBusiness, activatedBusiness any
	for _, write := range fake.Writes() {
		switch {
		case strings.Contains(write.query, "INSERT INTO user_provisioning_states"):
			// UpsertProvisioningState(user_id, business_id, mode, state, ...)
			provisioningBusiness = argAt(write.args, 1)
		case strings.Contains(write.query, "UPDATE authorization_business_members"):
			// ActivateNormalizedAuthorization(business_id, user_id)
			activatedBusiness = argAt(write.args, 0)
		}
	}
	if provisioningBusiness != int64(inviteBusinessID) {
		t.Fatalf("provisioning state written under business %v, want the invitation's business %d\n%s",
			provisioningBusiness, inviteBusinessID, statementSummary(fake.Writes()))
	}
	if activatedBusiness != int64(inviteBusinessID) {
		t.Fatalf("membership activated under business %v, want the invitation's business %d\n%s",
			activatedBusiness, inviteBusinessID, statementSummary(fake.Writes()))
	}
	for _, write := range fake.Writes() {
		for _, arg := range write.args {
			if id, ok := arg.(int64); ok && id == otherBusinessID {
				t.Fatalf("write carried the unrelated business %d:\n%s", id, statementSummary(fake.Writes()))
			}
		}
	}
}

// The owner lookup fails — a token whose owner owns no business row — and the
// member's users.business_id points at a real business. Nothing may be written.
func TestAcceptInviteRefusesWhenOwnerLookupFails(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		// business_id is NULL: the row predates the column being populated.
		inviteRowScript(nil),
		noBusinessForOwner,
	})
	service := NewService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "legacy-token", acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite succeeded with no business the invitation names, want error")
	}

	if writes := fake.Writes(); len(writes) != 0 {
		t.Fatalf("refused accept still wrote %d statement(s):\n%s", len(writes), statementSummary(writes))
	}
}

// The invitation names a business its issuer does not own. The accept must
// refuse rather than activate a member under an unrelated tenant.
func TestAcceptInviteRefusesBusinessOutsideInviterOwnership(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		businessInfoScript(otherBusinessID),
	})
	service := NewService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "stale-token", acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite activated a member in a business they have no membership in, want error")
	}

	if writes := fake.Writes(); len(writes) != 0 {
		t.Fatalf("refused accept still wrote %d statement(s):\n%s", len(writes), statementSummary(writes))
	}
}

// The password must not be set on the refusal path either: the token is only
// claimed after the tenant is verified, so a refused accept leaves the
// invitation and the account exactly as they were.
func TestAcceptInviteDoesNotTouchTheAccountOnRefusal(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		businessInfoScript(otherBusinessID),
	})
	service := NewService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "stale-token", acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite succeeded for an unverified business, want error")
	}

	for _, query := range fake.Queries() {
		if strings.Contains(query.query, "UPDATE users") || strings.Contains(query.query, "UPDATE invite_tokens") {
			t.Fatalf("refused accept mutated account or token state:\n%s", statementSummary(fake.Queries()))
		}
	}
}
