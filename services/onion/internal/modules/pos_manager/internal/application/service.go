package application

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/infrastructure"
)

type repository interface {
	VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.VerifyResult, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.VerifyResult, error) {
	return s.repo.VerifyPIN(ctx, businessID, pin)
}
