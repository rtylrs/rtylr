package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

func TestContainsNameMention(t *testing.T) {
	cases := []struct {
		content string
		name    string
		want    bool
	}{
		{"@mohamad issa ready for your check", "mohamad issa", true},
		{"thanks @mohamad issa.", "mohamad issa", true},
		{"ping @mohamad issa", "mohamad issa", true},
		{"@sam please look", "samira", false},
		{"@samira please look", "sam", false},
		{"mail sam@sam.io", "sam.io", false},
		{"شكرا @أحمد حمصي", "أحمد حمصي", true},
		{"@أحمد حمصيين", "أحمد حمصي", false},
		{"no mention of mohamad issa", "mohamad issa", false},
		{"@sam @samira", "samira", true},
		{"@anyone", "", false},
	}
	for _, tc := range cases {
		if got := containsNameMention(tc.content, tc.name); got != tc.want {
			t.Errorf("containsNameMention(%q, %q) = %v, want %v", tc.content, tc.name, got, tc.want)
		}
	}
}

// The comment box suggests people by display name and inserts "@Full Name".
// Most accounts have no username, so the mention must reach a project member
// through their name, or picking someone from the list notifies nobody.
func TestAddCommentNotifiesProjectMemberMentionedByName(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := NewService(ctx, testsupport.NewRepository(db))
	mentionID, _ := createNotificationUser(ctx, t, db)
	fullName := fmt.Sprintf("Mention Person %d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, "UPDATE users SET username = NULL, full_name = ? WHERE id = ?", fullName, mentionID); err != nil {
		t.Fatalf("rename mention user: %v", err)
	}
	task := seedTask(ctx, t, svc, "name mention task", cacheTestUserB)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: task.ProjectID, UserID: mentionID, Role: domain.FlowRoleMember,
	}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}

	if _, err := svc.AddComment(ctx, task.ID, cacheTestUserA, "@"+fullName+" ready for your check"); err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	mentioned, _, err := svc.ListNotifications(ctx, mentionID, 100, 0)
	if err != nil {
		t.Fatalf("ListNotifications: %v", err)
	}
	if _, found := findNotification(mentioned, mentionNotification, task.ID); !found {
		t.Fatal("a project member mentioned by full name received no mention notification")
	}
}
