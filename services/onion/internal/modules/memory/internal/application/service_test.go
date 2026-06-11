package application

import (
	"context"
	"strings"
	"testing"
)

func TestListUserSessions_RequiresUserID(t *testing.T) {
	s := &Service{}
	for _, uid := range []int{0, -1, -100} {
		if _, err := s.ListUserSessions(context.Background(), uid, 10, 0); err == nil {
			t.Fatalf("expected error for user id %d", uid)
		}
	}
}

func TestSetSessionMetadataFlagPreservesExistingMetadata(t *testing.T) {
	got, err := setSessionMetadataFlag(`{"archived":true,"theme":"dark"}`, "pinned", true)
	if err != nil {
		t.Fatalf("setSessionMetadataFlag: %v", err)
	}

	wantContains := []string{`"archived":true`, `"theme":"dark"`, `"pinned":true`}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata %s does not contain %s", got, want)
		}
	}
}

func TestSetSessionMetadataFlagInitializesEmptyMetadata(t *testing.T) {
	got, err := setSessionMetadataFlag("", "pinned", true)
	if err != nil {
		t.Fatalf("setSessionMetadataFlag: %v", err)
	}
	if got != `{"pinned":true}` {
		t.Fatalf("metadata = %s, want pinned true object", got)
	}
}
