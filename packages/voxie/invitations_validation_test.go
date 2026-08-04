package voxie_test

import (
	"testing"

	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
)

func TestInviteUserRequestAllowsEmailInviteWithoutPassword(t *testing.T) {
	req := &hubteaminvitationsv1.InviteUserRequest{
		Email:    "invitee@example.com",
		FullName: "Invitee",
		Role:     "staff",
		Apps:     []string{"dash"},
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("email invitation without password should be valid: %v", err)
	}
}

func TestAcceptInviteRequestRequiresPassword(t *testing.T) {
	req := &hubteaminvitationsv1.AcceptInviteRequest{Token: "invite-token"}

	if err := req.Validate(); err == nil {
		t.Fatal("accepting an invitation without a password should be invalid")
	}
}
