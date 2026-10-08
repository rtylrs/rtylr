package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/application"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	hubmemoryv1 "github.com/voxire/voxie/gen/go/hub/memory/v1"
)

type chatSessionLookupErrorRepository struct {
	*infrastructure.Repository
	err error
}

func (r chatSessionLookupErrorRepository) GetChatSession(context.Context, string, int) (*domain.ChatSession, error) {
	return nil, r.err
}

func TestGetChatMessagesReturnsNotFoundForUnownedSession(t *testing.T) {
	repository := chatSessionLookupErrorRepository{
		Repository: infrastructure.NewRepository(nil),
		err:        application.ErrChatSessionNotFound,
	}
	service := application.NewService(repository)
	api := NewMemoryAPI(service)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 7, BusinessID: 1})

	_, err := api.GetChatMessages(ctx, &hubmemoryv1.GetChatMessagesRequest{SessionId: "missing"})
	if err == nil {
		t.Fatal("expected a not-found error")
	}

	var httpError interface{ HTTPStatusCode() int }
	if !errors.As(err, &httpError) {
		t.Fatalf("error %T does not expose an HTTP status", err)
	}
	if got := httpError.HTTPStatusCode(); got != http.StatusNotFound {
		t.Fatalf("HTTP status = %d, want %d", got, http.StatusNotFound)
	}
}
