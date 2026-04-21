package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
)

// Service is the read-only business layer for /insights/*.
//
// It exists as a thin orchestrator: validate inputs, default date ranges,
// call the repository, and route through the platform cache. No SQL here.
type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}
