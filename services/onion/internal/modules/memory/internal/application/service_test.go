package application

import (
	"context"
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
