package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func workspaceInternalError(ctx context.Context, operation string, err error) error {
	logger.Ctx(ctx).Error(operation, logger.Error(err))
	return onekiterrors.New(http.StatusInternalServerError, "Internal server error")
}
