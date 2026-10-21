package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestActorPayloadMappersIncludeIdentity(t *testing.T) {
	actorID := int64(73)
	created := domain.TaskComment{
		ID:            1,
		TaskID:        2,
		UserID:        actorID,
		ActorName:     "Ada Lovelace",
		ActorInitials: "AL",
	}
	commentV1 := toComment(created)
	if commentV1.ActorName != "Ada Lovelace" || commentV1.ActorInitials != "AL" {
		t.Fatalf("v1 comment actor = %q/%q, want Ada Lovelace/AL", commentV1.ActorName, commentV1.ActorInitials)
	}
	commentV2 := toCommentV2(created)
	if commentV2.ActorName != "Ada Lovelace" || commentV2.ActorInitials != "AL" {
		t.Fatalf("v2 comment actor = %q/%q, want Ada Lovelace/AL", commentV2.ActorName, commentV2.ActorInitials)
	}

	activity := domain.TaskActivity{
		ID:            3,
		TaskID:        2,
		UserID:        actorID,
		ActorName:     "Ada Lovelace",
		ActorInitials: "AL",
	}
	activityV1 := toActivity(activity)
	if activityV1.ActorName != "Ada Lovelace" || activityV1.ActorInitials != "AL" {
		t.Fatalf("v1 activity actor = %q/%q, want Ada Lovelace/AL", activityV1.ActorName, activityV1.ActorInitials)
	}
	activityV2 := toActivityV2(activity)
	if activityV2.ActorName != "Ada Lovelace" || activityV2.ActorInitials != "AL" {
		t.Fatalf("v2 activity actor = %q/%q, want Ada Lovelace/AL", activityV2.ActorName, activityV2.ActorInitials)
	}

	notification := toNotification(domain.Notification{
		ID:            4,
		ActorID:       &actorID,
		ActorName:     "Ada Lovelace",
		ActorInitials: "AL",
	})
	if notification.GetActorId() != actorID || notification.ActorName != "Ada Lovelace" || notification.ActorInitials != "AL" {
		t.Fatalf("notification actor = %d/%q/%q, want %d/Ada Lovelace/AL", notification.GetActorId(), notification.ActorName, notification.ActorInitials, actorID)
	}
}
