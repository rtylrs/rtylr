package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) EvaluateFeature(ctx context.Context, featureKey string, userContext *domain.FeatureContext) (*domain.EvaluationResult, error) {
	flag, err := s.repo.GetFeatureFlag(ctx, featureKey)
	if err != nil {
		return s.createAndLogResult(ctx, featureKey, false, "feature_not_found", nil, userContext), err
	}

	if !flag.Enabled {
		return s.createAndLogResult(ctx, featureKey, false, "feature_disabled", nil, userContext), nil
	}

	if result := s.checkUserOverride(ctx, featureKey, userContext); result != nil {
		return result, nil
	}

	targetFlag, found, err := s.findTargetFlag(ctx, featureKey)
	if err != nil {
		return nil, err
	}

	if !found {
		return s.createAndLogResult(ctx, featureKey, false, "feature_not_found_with_rules", nil, userContext), nil
	}

	if len(targetFlag.Rules) == 0 {
		return s.evaluateDefaultRollout(ctx, featureKey, targetFlag, userContext), nil
	}

	if result := s.evaluateRules(ctx, featureKey, targetFlag, userContext); result != nil {
		return result, nil
	}

	return s.evaluateDefaultRollout(ctx, featureKey, targetFlag, userContext), nil
}

func (s *Service) EvaluateAllFeatures(ctx context.Context, userContext *domain.FeatureContext) (map[string]bool, error) {
	flagsWithRules, err := s.repo.GetFeatureFlagsWithRules(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string]bool)

	for _, flag := range flagsWithRules {
		result, err := s.EvaluateFeature(ctx, flag.FeatureFlag.Key, userContext)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to evaluate feature",
				logger.String("feature_key", flag.FeatureFlag.Key),
				logger.Error(err))
			results[flag.FeatureFlag.Key] = false

			continue
		}

		results[flag.FeatureFlag.Key] = result.Enabled
	}

	return results, nil
}

func (s *Service) evaluateRollout(userContext *domain.FeatureContext, percentage int32) bool {
	if percentage == 0 {
		return false
	}

	if percentage == 100 {
		return true
	}

	var seed int32
	if userContext != nil && !userContext.IsAnonymous {
		seed = userContext.UserID
	} else if userContext != nil && userContext.SessionID != "" {
		hash := 0
		for _, c := range userContext.SessionID {
			hash = hash*31 + int(c)
		}

		seed = int32(hash)
	} else {
		return false
	}

	return (seed % 100) < percentage
}

func (s *Service) createAndLogResult(ctx context.Context, featureKey string, enabled bool, reason string, ruleID *int32, userContext *domain.FeatureContext) *domain.EvaluationResult {
	result := &domain.EvaluationResult{
		FeatureKey: featureKey,
		Enabled:    enabled,
		Reason:     reason,
		RuleID:     ruleID,
		Context:    userContext,
	}
	s.logEvaluation(ctx, result)

	return result
}

func (s *Service) checkUserOverride(ctx context.Context, featureKey string, userContext *domain.FeatureContext) *domain.EvaluationResult {
	if userContext == nil || userContext.IsAnonymous {
		return nil
	}

	override, err := s.repo.GetUserFeatureOverride(ctx, userContext.UserID, featureKey)
	if err != nil {
		return nil
	}

	return s.createAndLogResult(ctx, featureKey, override.Enabled, "user_override: "+override.Reason, nil, userContext)
}

func (s *Service) findTargetFlag(ctx context.Context, featureKey string) (*domain.FeatureFlagWithRules, bool, error) {
	flagsWithRules, err := s.repo.GetFeatureFlagsWithRules(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get feature flags with rules", logger.Error(err))

		return nil, false, err
	}

	for _, f := range flagsWithRules {
		if f.FeatureFlag.Key == featureKey {
			return &f, true, nil
		}
	}

	return nil, false, nil
}

func (s *Service) evaluateDefaultRollout(ctx context.Context, featureKey string, targetFlag *domain.FeatureFlagWithRules, userContext *domain.FeatureContext) *domain.EvaluationResult {
	enabled := s.evaluateRollout(userContext, targetFlag.FeatureFlag.RolloutPercentage)

	reason := fmt.Sprintf("rollout_percentage_%d", targetFlag.FeatureFlag.RolloutPercentage)
	if len(targetFlag.Rules) > 0 {
		reason = fmt.Sprintf("default_rollout_%d", targetFlag.FeatureFlag.RolloutPercentage)
	}

	return s.createAndLogResult(ctx, featureKey, enabled, reason, nil, userContext)
}

func (s *Service) evaluateRules(ctx context.Context, featureKey string, targetFlag *domain.FeatureFlagWithRules, userContext *domain.FeatureContext) *domain.EvaluationResult {
	for _, rule := range targetFlag.Rules {
		if !rule.Enabled {
			continue
		}

		if s.EvaluateRule(userContext, &rule) {
			enabled := s.evaluateRuleAction(userContext, rule.Action, rule.RolloutPercentage)
			reason := fmt.Sprintf("rule_%s_%s", rule.Name, rule.Action)

			return s.createAndLogResult(ctx, featureKey, enabled, reason, &rule.Id, userContext)
		}
	}

	return nil
}

func (s *Service) evaluateRuleAction(userContext *domain.FeatureContext, action string, rolloutPercentage int32) bool {
	switch action {
	case "enable":
		return true
	case "disable":
		return false
	case "rollout":
		return s.evaluateRollout(userContext, rolloutPercentage)
	default:
		return false
	}
}

func (s *Service) logEvaluation(ctx context.Context, result *domain.EvaluationResult) {
	evaluation := domain.FeatureEvaluation{
		FeatureKey:  result.FeatureKey,
		Enabled:     result.Enabled,
		Reason:      result.Reason,
		RuleID:      result.RuleID,
		UserContext: result.Context,
		EvaluatedAt: time.Now(),
	}

	if result.Context != nil && !result.Context.IsAnonymous {
		evaluation.UserID = &result.Context.UserID
	}

	s.repo.LogFeatureEvaluation(ctx, evaluation)
}
