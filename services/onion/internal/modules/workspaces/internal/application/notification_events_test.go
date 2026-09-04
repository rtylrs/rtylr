package application

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestMentionedUsernames(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    []string
	}{
		{name: "none", content: "plain comment", want: []string{}},
		{name: "one", content: "Please check @flow.member", want: []string{"flow.member"}},
		{name: "deduplicates", content: "@Alice and @alice", want: []string{"alice"}},
		{name: "ignores email", content: "mail alice@example.com and greet @bob", want: []string{"bob"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mentionedUsernames(tc.content)
			if len(got) != len(tc.want) {
				t.Fatalf("mentionedUsernames(%q) = %v, want %v", tc.content, got, tc.want)
			}
			for index := range tc.want {
				if got[index] != tc.want[index] {
					t.Errorf("mentionedUsernames(%q)[%d] = %q, want %q", tc.content, index, got[index], tc.want[index])
				}
			}
		})
	}
}

func TestAddCommentEmitsCommentAndMentionNotifications(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := NewService(ctx, db)
	mentionID, username := createNotificationUser(ctx, t, db)
	task := seedTask(ctx, t, svc, "notification task", cacheTestUserB)

	comment, err := svc.AddComment(ctx, task.ID, cacheTestUserA, "Please review this @"+username)
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if comment.CreatedAt.IsZero() || comment.UpdatedAt.IsZero() {
		t.Fatalf("comment timestamps are empty: created_at=%v updated_at=%v", comment.CreatedAt, comment.UpdatedAt)
	}

	notifications, _, err := svc.ListNotifications(ctx, cacheTestUserB, 100, 0)
	if err != nil {
		t.Fatalf("ListNotifications for assignee: %v", err)
	}
	if notification, found := findNotification(notifications, commentNotification, task.ID); !found {
		t.Fatal("assignee did not receive the comment notification")
	} else {
		assertNotificationActor(t, notification, cacheTestUserA)
	}

	mentioned, _, err := svc.ListNotifications(ctx, mentionID, 100, 0)
	if err != nil {
		t.Fatalf("ListNotifications for mention: %v", err)
	}
	if notification, found := findNotification(mentioned, mentionNotification, task.ID); !found {
		t.Fatal("mentioned user did not receive the mention notification")
	} else {
		assertNotificationActor(t, notification, cacheTestUserA)
	}
}

func TestTaskStatusChangeHonorsNotificationSetting(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := NewService(ctx, db)
	task := seedTask(ctx, t, svc, "status notification task", cacheTestUserB)

	if err := svc.UpdateUserSettings(ctx, domain.UserSettings{
		UserID:        cacheTestUserB,
		TaskAssigned:  true,
		TaskComments:  true,
		Mentions:      true,
		StatusChanges: true,
		WeeklyDigest:  true,
	}); err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}

	if err := svc.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Status:     domain.TaskStatusInProgress,
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	notifications, _, err := svc.ListNotifications(ctx, cacheTestUserB, 100, 0)
	if err != nil {
		t.Fatalf("ListNotifications: %v", err)
	}
	if notification, found := findNotification(notifications, statusNotification, task.ID); !found {
		t.Fatal("assignee did not receive the status change notification")
	} else {
		assertNotificationActor(t, notification, cacheTestUserA)
		if notification.Description != domain.TaskStatusInProgress {
			t.Errorf("status notification description = %q, want %q", notification.Description, domain.TaskStatusInProgress)
		}
	}
}

func findNotification(notifications []domain.Notification, notificationType string, taskID int64) (domain.Notification, bool) {
	for _, notification := range notifications {
		if notification.Type == notificationType && notification.TaskID != nil && *notification.TaskID == taskID {
			return notification, true
		}
	}
	return domain.Notification{}, false
}

func assertNotificationActor(t *testing.T, notification domain.Notification, actorID int64) {
	t.Helper()
	if notification.ActorID == nil || *notification.ActorID != actorID {
		t.Errorf("notification actor = %v, want %d", notification.ActorID, actorID)
	}
}

func createNotificationUser(ctx context.Context, t *testing.T, db *sql.DB) (int64, string) {
	t.Helper()
	username := fmt.Sprintf("notify_%d", time.Now().UnixNano())
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (business_id, email, username, password_hash, full_name, onboarded) VALUES (?, ?, ?, ?, ?, TRUE)",
		cacheTestBusiness, username+"@test.local", username, "test-password-hash", "Notification User")
	if err != nil {
		t.Fatalf("create notification user: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("notification user id: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_notifications WHERE user_id = ?", id); err != nil {
			t.Errorf("delete notification rows: %v", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_user_settings WHERE user_id = ?", id); err != nil {
			t.Errorf("delete notification settings: %v", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id); err != nil {
			t.Errorf("delete notification user: %v", err)
		}
	})
	return id, username
}
