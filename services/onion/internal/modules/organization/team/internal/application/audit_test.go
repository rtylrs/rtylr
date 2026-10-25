package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
)

type recordingAudit struct {
	events []audit.Event
	err    error
}

func (r *recordingAudit) RecordAudit(_ context.Context, event audit.Event) error {
	r.events = append(r.events, event)
	return r.err
}

// groupRepo answers only the calls the group mutations make; anything else
// panics on the nil embedded interface, which flags an unexpected call.
type groupRepo struct {
	repository
	group     *domain.Group
	deleteErr error
}

func (g *groupRepo) GetGroup(context.Context, int64, int64) (*domain.Group, error) {
	return g.group, nil
}

func (g *groupRepo) GetGroupDetail(context.Context, int64, int64) (*domain.GroupDetail, error) {
	return nil, errors.New("not needed")
}

// The actor below is the owner, so delegation is unrestricted.
func (g *groupRepo) GetBusinessByID(context.Context, int64) (string, int64, error) {
	return "Biz", 1, nil
}

func (g *groupRepo) DeleteGroup(context.Context, int64, int64) error { return g.deleteErr }

func (g *groupRepo) RemoveGroupMember(context.Context, int64, int64, int64) error { return nil }

func TestDeleteGroup_RecordsAuditEntryWithGroupName(t *testing.T) {
	recorder := &recordingAudit{}
	s := NewService(context.Background(), &groupRepo{group: &domain.Group{ID: 7, Name: "Kitchen"}}).WithAuditRecorder(recorder)

	if err := s.DeleteGroup(context.Background(), 1, 42, 7); err != nil {
		t.Fatalf("DeleteGroup: %v", err)
	}
	if len(recorder.events) != 1 {
		t.Fatalf("recorded %d audit events, want 1", len(recorder.events))
	}
	got := recorder.events[0]
	want := audit.Event{BusinessID: 42, EntityType: "group", EntityID: 7, Action: "group.deleted", Description: `Deleted group "Kitchen"`}
	if got != want {
		t.Fatalf("audit event = %+v, want %+v", got, want)
	}
}

func TestDeleteGroup_FailedChangeRecordsNothing(t *testing.T) {
	recorder := &recordingAudit{}
	s := NewService(context.Background(), &groupRepo{group: &domain.Group{ID: 7, Name: "Kitchen"}, deleteErr: errors.New("boom")}).WithAuditRecorder(recorder)

	if err := s.DeleteGroup(context.Background(), 1, 42, 7); err == nil {
		t.Fatal("DeleteGroup succeeded, want the repository error")
	}
	if len(recorder.events) != 0 {
		t.Fatalf("recorded %d audit events for a change that failed, want 0", len(recorder.events))
	}
}

// The change has already committed by the time it is audited, so a recorder
// failure must not turn a successful removal into an error.
func TestRemoveGroupMember_AuditFailureDoesNotFailTheChange(t *testing.T) {
	recorder := &recordingAudit{err: errors.New("outbox down")}
	s := NewService(context.Background(), &groupRepo{group: &domain.Group{ID: 7, Name: "Kitchen"}}).WithAuditRecorder(recorder)

	if err := s.RemoveGroupMember(context.Background(), 1, 42, 7, 99); err != nil {
		t.Fatalf("RemoveGroupMember: %v", err)
	}
	if len(recorder.events) != 1 || recorder.events[0].Action != "group.member_removed" {
		t.Fatalf("audit events = %+v, want one group.member_removed", recorder.events)
	}
}

func TestRecordAudit_SkipsIncompleteIdentity(t *testing.T) {
	recorder := &recordingAudit{}
	s := (&Service{}).WithAuditRecorder(recorder)

	s.recordAudit(context.Background(), 0, 1, "member", 5, "member.invited", "")
	s.recordAudit(context.Background(), 42, 1, "member", 0, "member.invited", "")
	if len(recorder.events) != 0 {
		t.Fatalf("recorded %d events without a business or entity, want 0", len(recorder.events))
	}
}

func TestMemberLabel(t *testing.T) {
	cases := map[[2]string]string{
		{"Jane Doe", "jane@example.com"}: "Jane Doe (jane@example.com)",
		{"  Jane  ", ""}:                 "Jane",
		{"", "jane@example.com"}:         "jane@example.com",
		{"", ""}:                         "a member",
	}
	for in, want := range cases {
		if got := memberLabel(in[0], in[1]); got != want {
			t.Errorf("memberLabel(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}
