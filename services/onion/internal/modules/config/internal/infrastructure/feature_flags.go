package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) GetFeatureFlag(ctx context.Context, key string) (domain.FeatureFlag, error) {
	return db.NewReadExecutor[domain.FeatureFlag, sqlc.GetFeatureFlagRow](db.MainDatabase).Fetch(ctx, r.queries.GetFeatureFlag, key)
}

func (r *Repository) GetAllFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error) {
	return db.NewReadExecutor[domain.FeatureFlag, sqlc.GetAllFeatureFlagsRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetAllFeatureFlags)
}

func (r *Repository) GetEnabledFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error) {
	return db.NewReadExecutor[domain.FeatureFlag, sqlc.GetEnabledFeatureFlagsRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetEnabledFeatureFlags)
}

func (r *Repository) GetFeatureFlagsWithRules(ctx context.Context) ([]domain.FeatureFlagWithRules, error) {
	rows, err := db.NewReadExecutor[domain.FeatureFlagWithRulesRow, sqlc.GetFeatureFlagsWithRulesRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetFeatureFlagsWithRules)
	if err != nil {
		return nil, err
	}

	return domain.GroupFeatureFlagsWithRules(rows), nil
}

// User Feature Override methods

func (r *Repository) GetUserFeatureOverrides(ctx context.Context, userID int32) ([]domain.UserFeatureOverride, error) {
	return db.NewReadExecutor[domain.UserFeatureOverride, sqlc.GetUserFeatureOverridesRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetUserFeatureOverrides, userID)
}

func (r *Repository) GetUserFeatureOverride(ctx context.Context, userID int32, featureKey string) (domain.UserFeatureOverride, error) {
	params := sqlc.GetUserFeatureOverrideParams{
		UserID:     userID,
		FeatureKey: featureKey,
	}

	return db.NewReadExecutor[domain.UserFeatureOverride, sqlc.GetUserFeatureOverrideRow](db.MainDatabase).Fetch(ctx, r.queries.GetUserFeatureOverride, params)
}

func (r *Repository) LogFeatureEvaluation(ctx context.Context, evaluation domain.FeatureEvaluation) {
	var (
		userContext []byte
		err         error
	)

	if evaluation.UserContext != nil {
		userContext, err = evaluation.UserContext.ToJSON()
		if err != nil {
			logger.Ctx(ctx).Error("Failed to marshal user context", logger.Error(err))

			return
		}
	}

	var userID int32
	if evaluation.UserID != nil {
		userID = *evaluation.UserID
	}

	var ruleID int32
	if evaluation.RuleID != nil {
		ruleID = *evaluation.RuleID
	}

	params := sqlc.LogFeatureEvaluationParams{
		UserID:      sql.NullInt32{Int32: userID, Valid: evaluation.UserID != nil},
		FeatureKey:  evaluation.FeatureKey,
		Enabled:     evaluation.Enabled,
		Reason:      sql.NullString{String: evaluation.Reason, Valid: evaluation.Reason != ""},
		RuleID:      sql.NullInt32{Int32: ruleID, Valid: evaluation.RuleID != nil},
		UserContext: userContext,
	}

	if err := r.queries.LogFeatureEvaluation(ctx, params); err != nil {
		logger.Ctx(ctx).Error("Failed to log feature evaluation", logger.Error(err))
	}
}
