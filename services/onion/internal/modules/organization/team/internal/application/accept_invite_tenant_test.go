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
	newOwnerID       int64 = 7005
)

func inviteRowScript(inviteBusiness int64) scriptedQuery {
	return scriptedQuery{
		match:   containsAny("FROM invite_tokens"),
		columns: []string{"user_id", "business_id", "expires_at", "used_at"},
		rows: [][]driver.Value{{
			inviteeID, inviteBusiness, time.Now().Add(time.Hour), nil,
		}},
	}
}

func businessInfoScript(ownerID int64) scriptedQuery {
	return scriptedQuery{
		match:   containsAny("SELECT name, owner_id FROM businesses WHERE id = ?"),
		columns: []string{"name", "owner_id"},
		rows:    [][]driver.Value{{"Invited Business", ownerID}},
	}
}

// The invited business no longer exists.
var missingBusiness = scriptedQuery{
	match:   containsAny("SELECT name, owner_id FROM businesses WHERE id = ?"),
	columns: []string{"name", "owner_id"},
	err:     sql.ErrNoRows,
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

// acceptedBusinesses reports the business the accept's two tenant writes used.
func acceptedBusinesses(fake *fakeDatabase) (provisioning, activated any) {
	for _, write := range fake.Writes() {
		switch {
		case strings.Contains(write.query, "INSERT INTO user_provisioning_states"):
			// UpsertProvisioningState(user_id, business_id, mode, state, ...)
			provisioning = argAt(write.args, 1)
		case strings.Contains(write.query, "UPDATE authorization_business_members"):
			// ActivateNormalizedAuthorization(business_id, user_id)
			activated = argAt(write.args, 0)
		}
	}
	return provisioning, activated
}

// A member who already belongs to one business is invited to another. The
// accept uses the invitation's business.
func TestAcceptInviteWritesUnderTheInviteBusiness(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		businessInfoScript(inviteOwnerID),
	})
	service := newTestService(context.Background(), db)

	userID, err := service.AcceptInvite(context.Background(), "live-token", acceptInvitePassword)
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if userID != inviteeID {
		t.Fatalf("AcceptInvite returned user %d, want %d", userID, inviteeID)
	}

	provisioningBusiness, activatedBusiness := acceptedBusinesses(fake)
	if provisioningBusiness != inviteBusinessID {
		t.Fatalf("provisioning state written under business %v, want the invitation's business %d\n%s",
			provisioningBusiness, inviteBusinessID, statementSummary(fake.Writes()))
	}
	if activatedBusiness != inviteBusinessID {
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

// The business changed hands after the invitation went out. The invitation
// belongs to the business, so it still joins the member there.
func TestAcceptInviteSurvivesAnOwnershipTransfer(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		businessInfoScript(newOwnerID),
	})
	service := newTestService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "live-token", acceptInvitePassword); err != nil {
		t.Fatalf("AcceptInvite after an ownership transfer: %v", err)
	}
	if _, activated := acceptedBusinesses(fake); activated != inviteBusinessID {
		t.Fatalf("membership activated under business %v, want %d\n%s",
			activated, inviteBusinessID, statementSummary(fake.Writes()))
	}
}

// An invitation that names no business must write nothing, even though the
// member's users.business_id may point at a real business.
func TestAcceptInviteRefusesInvitationWithoutBusiness(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{inviteRowScript(0)})
	service := newTestService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "legacy-token", acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite succeeded with no business the invitation names, want error")
	}
	if writes := fake.Writes(); len(writes) != 0 {
		t.Fatalf("refused accept still wrote %d statement(s):\n%s", len(writes), statementSummary(writes))
	}
}

// The password must not be set on the refusal path either: the token is only
// claimed after the tenant is verified, so a refused accept leaves the
// invitation and the account exactly as they were.
func TestAcceptInviteDoesNotTouchTheAccountWhenTheBusinessIsGone(t *testing.T) {
	db, fake := newFakeDB(t, []scriptedQuery{
		inviteRowScript(inviteBusinessID),
		missingBusiness,
	})
	service := newTestService(context.Background(), db)

	if _, err := service.AcceptInvite(context.Background(), "stale-token", acceptInvitePassword); err == nil {
		t.Fatal("AcceptInvite succeeded for a business that no longer exists, want error")
	}
	for _, query := range fake.Queries() {
		if strings.Contains(query.query, "UPDATE users") || strings.Contains(query.query, "UPDATE invite_tokens") {
			t.Fatalf("refused accept mutated account or token state:\n%s", statementSummary(fake.Queries()))
		}
	}
}
