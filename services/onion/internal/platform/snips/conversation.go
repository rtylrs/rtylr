package snips

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

type conversationKey struct{}

// WithConversation keeps provider routing stable across turns without exposing
// the internal conversation identifier to the provider.
func WithConversation(ctx context.Context, id string) context.Context {
	if id == "" {
		return ensureConversation(ctx)
	}
	digest := sha256.Sum256([]byte("rtylr-caret:" + id))
	return context.WithValue(ctx, conversationKey{}, hex.EncodeToString(digest[:]))
}

func ensureConversation(ctx context.Context) context.Context {
	if id, _ := ctx.Value(conversationKey{}).(string); id != "" {
		return ctx
	}
	return context.WithValue(ctx, conversationKey{}, uuid.NewString())
}
