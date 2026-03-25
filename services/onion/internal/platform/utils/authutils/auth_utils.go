package authutils

import (
	"context"

	models "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/server/http"
)

func GetAuthUserFromContext(ctx context.Context) *models.AuthUser {
	user, ok := http.GetAuthUser(ctx)
	if !ok {
		return nil
	}

	return user
}

func IsAuthenticated(ctx context.Context) bool {
	_, ok := http.GetAuthUser(ctx)

	return ok
}
