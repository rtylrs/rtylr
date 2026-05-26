package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
)

func TestInviteUser_RequiresEmail(t *testing.T) {
	s := &Service{}
	if _, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{}); err == nil {
		t.Fatalf("expected error for missing email")
	}
}

func TestUpdatePermissions_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.UpdatePermissions(context.Background(), 1, 0, domain.UpdatePermissionsInput{}); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}

func TestDeactivateUser_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.DeactivateUser(context.Background(), 1, 0, 0); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}

func TestCreateGroup_RequiresName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateGroup(context.Background(), 1, domain.CreateGroupInput{Name: ""}); err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestCreateGroup_NameRequired_GivesError(t *testing.T) {
	s := &Service{}
	_, err := s.CreateGroup(context.Background(), 1, domain.CreateGroupInput{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestUpdateGroup_RequiresName(t *testing.T) {
	s := &Service{}
	if err := s.UpdateGroup(context.Background(), 1, 5, domain.UpdateGroupInput{Name: ""}); err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestAddGroupMembers_RequiresUserIDs(t *testing.T) {
	s := &Service{}
	if err := s.AddGroupMembers(context.Background(), 1, 5, 1, nil); err == nil {
		t.Fatal("expected error when no user IDs provided")
	}
}

func TestAddGroupMembers_RequiresNonEmptyUserIDs(t *testing.T) {
	s := &Service{}
	if err := s.AddGroupMembers(context.Background(), 1, 5, 1, []int64{}); err == nil {
		t.Fatal("expected error for empty user IDs slice")
	}
}

func TestBulkInvite_SkipsEmptyEmails(t *testing.T) {
	s := &Service{}
	results, _ := s.BulkInvite(context.Background(), 1, 0, domain.BulkInviteInput{
		Emails: []string{"", "  ", ""},
		Apps:   []string{"pos"},
		Role:   domain.RoleStaff,
	})
	// All emails are blank, so all should be skipped and results be nil/empty.
	if len(results) != 0 {
		t.Errorf("expected 0 results for all-empty emails, got %d", len(results))
	}
}

func TestInviteUser_EmptyEmailStringFails(t *testing.T) {
	// Service checks Email == "" before any repo call.
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{Email: ""})
	if err == nil {
		t.Fatal("expected error for empty email string")
	}
}

func TestPaginateTeam_DefaultsPage(t *testing.T) {
	limit, offset := paginateTeam(0, 10)
	if limit != 10 || offset != 0 {
		t.Errorf("page=0 should default to page 1: limit=%d offset=%d", limit, offset)
	}
}

func TestPaginateTeam_DefaultsPageSize(t *testing.T) {
	limit, offset := paginateTeam(1, 0)
	if limit != 50 || offset != 0 {
		t.Errorf("pageSize=0 should default to 50: limit=%d offset=%d", limit, offset)
	}
}

func TestPaginateTeam_CapsPageSize(t *testing.T) {
	limit, _ := paginateTeam(1, 9999)
	if limit != 500 {
		t.Errorf("pageSize should be capped at 500, got %d", limit)
	}
}

func TestPaginateTeam_Page2Offset(t *testing.T) {
	limit, offset := paginateTeam(2, 25)
	if limit != 25 || offset != 25 {
		t.Errorf("page=2 pageSize=25: limit=%d offset=%d", limit, offset)
	}
}

func TestUpdatePermissions_ErrorMessage(t *testing.T) {
	s := &Service{}
	err := s.UpdatePermissions(context.Background(), 1, 0, domain.UpdatePermissionsInput{UserID: 0})
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestDeactivateUser_ErrorMessage(t *testing.T) {
	s := &Service{}
	err := s.DeactivateUser(context.Background(), 1, 0, 0)
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestInviteUser_ErrorMessage(t *testing.T) {
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{})
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestGetBusinessInfo_ZeroBusinessID(t *testing.T) {
	s := &Service{}
	name, isOwner := s.GetBusinessInfo(context.Background(), 0, 1)
	if name != "" || isOwner {
		t.Errorf("expected empty name and not owner for businessID=0: got name=%q isOwner=%v", name, isOwner)
	}
}

func TestCreateGroup_EmptyNameAlwaysErrors(t *testing.T) {
	s := &Service{}
	_, err := s.CreateGroup(context.Background(), 0, domain.CreateGroupInput{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name with zero ownerID")
	}
}
