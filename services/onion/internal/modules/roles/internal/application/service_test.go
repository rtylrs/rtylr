package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

func TestCreateRole_RequiresName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateRole(context.Background(), 1, domain.CreateRoleInput{}); err == nil {
		t.Fatalf("expected error for missing name")
	}
}

func TestUpdateRole_RequiresID(t *testing.T) {
	s := &Service{}
	if err := s.UpdateRole(context.Background(), 1, domain.UpdateRoleInput{}); err == nil {
		t.Fatalf("expected error for missing role id")
	}
}

func TestDeleteRole_RequiresID(t *testing.T) {
	s := &Service{}
	if err := s.DeleteRole(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing role id")
	}
}
