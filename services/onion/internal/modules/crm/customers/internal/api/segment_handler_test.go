package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestDeleteSegmentHandler_RejectsNonManagerBeforeService(t *testing.T) {
	ctx := (&authPkg.Device{
		User: &authPkg.AuthUser{
			UserID:     7,
			BusinessID: 42,
		},
		IsAnonymous: false,
	}).Ctx(context.Background())

	h := &DeleteSegmentHandler{}
	if _, err := h.Serve(ctx, &DeleteSegmentRequest{ID: 11}); err == nil {
		t.Fatal("expected non-manager segment deletion to be rejected")
	}
}
