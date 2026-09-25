package application

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
)

// Service is the read-only business layer for /insights/*.
//
// It exists as a thin orchestrator: validate inputs, default date ranges,
// call the repository, and route through the platform cache. No SQL here.
type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// GetProductMarginRows returns per-product margin rows for margin reporting.
func (s *Service) GetProductMarginRows(ctx context.Context, businessID int64) ([]domain.ProductMarginRow, error) {
	return s.repo.GetProductMarginRows(ctx, businessID)
}
