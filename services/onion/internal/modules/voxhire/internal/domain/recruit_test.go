package domain

import (
	"context"
	"testing"
)

func TestOwnerIDContextRoundTrip(t *testing.T) {
	ctx := WithOwnerID(context.Background(), 42)

	ownerID, ok := OwnerIDFromContext(ctx)
	if !ok || ownerID != 42 {
		t.Fatalf("OwnerIDFromContext() = (%d, %v), want (42, true)", ownerID, ok)
	}
}

func TestOwnerIDContextRejectsMissingOrInvalidScope(t *testing.T) {
	for _, ctx := range []context.Context{
		context.Background(),
		WithOwnerID(context.Background(), 0),
		WithOwnerID(context.Background(), -1),
	} {
		if ownerID, ok := OwnerIDFromContext(ctx); ok || ownerID != 0 {
			t.Fatalf("OwnerIDFromContext() = (%d, %v), want (0, false)", ownerID, ok)
		}
	}
}
