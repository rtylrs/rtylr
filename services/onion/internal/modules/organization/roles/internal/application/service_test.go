package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
)

func TestAuthorizationServiceRequiresNormalizedRepository(t *testing.T) {
	service := &Service{}
	if _, err := service.ListPermissionCatalog(context.Background()); !errors.Is(err, domain.ErrAuthorizationSchemaUnavailable) {
		t.Fatalf("ListPermissionCatalog error = %v", err)
	}
	if _, err := service.ListManagedRoles(context.Background(), 1, domain.RoleFilter{}); !errors.Is(err, domain.ErrAuthorizationSchemaUnavailable) {
		t.Fatalf("ListManagedRoles error = %v", err)
	}
}

func TestCreateManagedRoleRequiresName(t *testing.T) {
	service := &Service{}
	if _, err := service.CreateManagedRole(context.Background(), 1, 2, domain.CreateRoleInput{}); err == nil {
		t.Fatal("expected a missing role name to be rejected")
	}
}

func TestUpdateManagedRoleRequiresIDAndName(t *testing.T) {
	service := &Service{}
	if _, err := service.UpdateManagedRole(context.Background(), 1, 2, domain.UpdateRoleInput{}); err == nil {
		t.Fatal("expected a missing role id to be rejected")
	}
	if _, err := service.UpdateManagedRole(context.Background(), 1, 2, domain.UpdateRoleInput{ID: 1}); err == nil {
		t.Fatal("expected a missing role name to be rejected")
	}
}

func TestAssignManagedRoleRequiresOneTarget(t *testing.T) {
	service := &Service{}
	_, err := service.AssignManagedRole(context.Background(), 1, 2, domain.AssignRoleInput{
		RoleID: 1,
		AppID:  "pos",
	})
	if !errors.Is(err, domain.ErrInvalidAssignmentTarget) {
		t.Fatalf("AssignManagedRole error = %v", err)
	}
}
