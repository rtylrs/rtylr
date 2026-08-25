package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) insightsCacheKey(ctx context.Context, businessID int64, detail string) (string, error) {
	generation, err := s.repo.GetInsightsGeneration(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: failed to resolve cache generation", logger.Error(err))
		return "", err
	}
	return fmt.Sprintf("insights:%d:g:%d:%s", businessID, generation.UnixNano(), detail), nil
}
