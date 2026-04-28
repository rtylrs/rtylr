package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
)

func TestInviteUser_RequiresEmail(t *testing.T) {
	s := &Service{}
	if _, _, err := s.InviteUser(context.Background(), 1, domain.InviteUserInput{}); err == nil {
		t.Fatalf("expected error for missing email")
	}
}

func TestUpdatePermissions_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.UpdatePermissions(context.Background(), 1, domain.UpdatePermissionsInput{}); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}

func TestDeactivateUser_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.DeactivateUser(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}
