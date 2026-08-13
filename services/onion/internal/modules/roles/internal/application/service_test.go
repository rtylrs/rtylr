package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/roles/internal/domain"
)

type mockRepo struct {
	listResult []domain.Role
	listErr    error
	createID   int64
	createErr  error
	updateErr  error
	deleteErr  error
	createGot  struct {
		ownerID int64
		name    string
		perms   []string
	}
	updateGot struct {
		ownerID, roleID int64
		name            string
		perms           []string
	}
	deleteGot struct {
		ownerID, roleID int64
	}
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Role, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, ownerID int64, name string, perms []string) (int64, error) {
	m.createGot.ownerID = ownerID
	m.createGot.name = name
	m.createGot.perms = perms
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, ownerID, roleID int64, name string, perms []string) error {
	m.updateGot.ownerID = ownerID
	m.updateGot.roleID = roleID
	m.updateGot.name = name
	m.updateGot.perms = perms
	return m.updateErr
}

func (m *mockRepo) Delete(_ context.Context, ownerID, roleID int64) error {
	m.deleteGot.ownerID = ownerID
	m.deleteGot.roleID = roleID
	return m.deleteErr
}

func TestListRoles_Delegates(t *testing.T) {
	want := []domain.Role{{ID: 1, Name: "admin"}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.ListRoles(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].Name != "admin" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestCreateRole_NilPermsBecomeEmpty(t *testing.T) {
	m := &mockRepo{createID: 5}
	s := &Service{repo: m}
	if _, err := s.CreateRole(context.Background(), 1, domain.CreateRoleInput{Name: "r"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.createGot.perms == nil || len(m.createGot.perms) != 0 {
		t.Errorf("expected empty perms, got %v", m.createGot.perms)
	}
}

func TestCreateRole_PassesPerms(t *testing.T) {
	m := &mockRepo{createID: 5}
	s := &Service{repo: m}
	if _, err := s.CreateRole(context.Background(), 1, domain.CreateRoleInput{Name: "r", Permissions: []string{"a", "b"}}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(m.createGot.perms) != 2 || m.createGot.perms[0] != "a" {
		t.Errorf("perms: %v", m.createGot.perms)
	}
}

func TestCreateRole_PropagatesError(t *testing.T) {
	want := errors.New("create failed")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateRole(context.Background(), 1, domain.CreateRoleInput{Name: "r"}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateRole_NilPermsBecomeEmpty(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.UpdateRole(context.Background(), 1, domain.UpdateRoleInput{ID: 7, Name: "x"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.updateGot.perms == nil || len(m.updateGot.perms) != 0 {
		t.Errorf("expected empty perms, got %v", m.updateGot.perms)
	}
	if m.updateGot.roleID != 7 {
		t.Errorf("roleID didn't pass through: %d", m.updateGot.roleID)
	}
}

func TestUpdateRole_PropagatesNotFound(t *testing.T) {
	m := &mockRepo{updateErr: domain.ErrRoleNotFound}
	s := &Service{repo: m}
	if err := s.UpdateRole(context.Background(), 1, domain.UpdateRoleInput{ID: 7, Name: "missing"}); !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestDeleteRole_DelegatesIDs(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.DeleteRole(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.deleteGot.ownerID != 7 || m.deleteGot.roleID != 42 {
		t.Errorf("delete args: %+v", m.deleteGot)
	}
}

func TestDeleteRole_PropagatesNotFound(t *testing.T) {
	m := &mockRepo{deleteErr: domain.ErrRoleNotFound}
	s := &Service{repo: m}
	if err := s.DeleteRole(context.Background(), 1, 7); !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestListPermissions_NonEmpty(t *testing.T) {
	s := &Service{}
	if got := s.ListPermissions(); len(got) == 0 {
		t.Error("ListPermissions should return at least one entry")
	}
}

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
