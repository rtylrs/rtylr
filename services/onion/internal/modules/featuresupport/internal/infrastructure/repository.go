package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), pool: db}
}

func flagFromRow(r sqlc.FeatureFlag) domain.Flag {
	return domain.Flag{
		Key:               r.FeatureKey,
		Description:       r.Description,
		Enabled:           r.Enabled,
		RolloutPercentage: r.RolloutPercentage,
		VariantOn:         r.VariantOn,
		VariantOff:        r.VariantOff,
	}
}

func targetFromRow(r sqlc.FeatureTarget) domain.Target {
	return domain.Target{
		FeatureKey: r.FeatureKey,
		OwnerID:    r.OwnerID,
		Enabled:    r.Enabled,
		Variant:    r.Variant.String,
	}
}

func (r *Repository) GetFlag(ctx context.Context, key string) (domain.Flag, error) {
	row, err := r.queries.GetFeatureFlag(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Flag{}, domain.ErrFlagNotFound
		}
		logger.Ctx(ctx).Error("Failed to get feature flag", logger.Error(err))
		return domain.Flag{}, err
	}
	return flagFromRow(row), nil
}

func (r *Repository) ListFlags(ctx context.Context) ([]domain.Flag, error) {
	rows, err := r.queries.ListFeatureFlags(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature flags", logger.Error(err))
		return nil, err
	}
	out := make([]domain.Flag, 0, len(rows))
	for _, row := range rows {
		out = append(out, flagFromRow(row))
	}
	return out, nil
}

func (r *Repository) UpsertFlag(ctx context.Context, f domain.Flag) error {
	err := r.queries.UpsertFeatureFlag(ctx, sqlc.UpsertFeatureFlagParams{
		FeatureKey:        f.Key,
		Description:       f.Description,
		Enabled:           f.Enabled,
		RolloutPercentage: f.RolloutPercentage,
		VariantOn:         f.VariantOn,
		VariantOff:        f.VariantOff,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert feature flag", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteFlag(ctx context.Context, key string) error {
	err := r.queries.DeleteFeatureFlag(ctx, key)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete feature flag", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUserEmail(ctx context.Context, userID int64) (string, error) {
	// cross-schema read from users — feature evaluation subject email
	email, err := r.queries.GetUserEmailByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		logger.Ctx(ctx).Error("Failed to get feature subject email", logger.Error(err))
		return "", err
	}
	return email, nil
}

func (r *Repository) ListTargetsByOwner(ctx context.Context, ownerID int64) ([]domain.Target, error) {
	rows, err := r.queries.ListFeatureTargetsByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature targets by owner", logger.Error(err))
		return nil, err
	}
	out := make([]domain.Target, 0, len(rows))
	for _, row := range rows {
		out = append(out, targetFromRow(row))
	}
	return out, nil
}

func (r *Repository) ListTargetsByKey(ctx context.Context, key string) ([]domain.Target, error) {
	rows, err := r.queries.ListFeatureTargetsByKey(ctx, key)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature targets by key", logger.Error(err))
		return nil, err
	}
	out := make([]domain.Target, 0, len(rows))
	for _, row := range rows {
		out = append(out, targetFromRow(row))
	}
	return out, nil
}

func (r *Repository) UpsertTarget(ctx context.Context, t domain.Target) error {
	err := r.queries.UpsertFeatureTarget(ctx, sqlc.UpsertFeatureTargetParams{
		FeatureKey: t.FeatureKey,
		OwnerID:    t.OwnerID,
		Enabled:    t.Enabled,
		Variant:    sql.NullString{String: t.Variant, Valid: t.Variant != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert feature target", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteTarget(ctx context.Context, key string, ownerID int64) error {
	err := r.queries.DeleteFeatureTarget(ctx, sqlc.DeleteFeatureTargetParams{
		FeatureKey: key,
		OwnerID:    ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete feature target", logger.Error(err))
	}
	return err
}

func ruleFromRow(r sqlc.FeatureRule) domain.Rule {
	return domain.Rule{
		ID:                r.ID,
		FeatureKey:        r.FeatureKey,
		Priority:          r.Priority,
		Description:       r.Description,
		Variant:           r.Variant,
		RolloutPercentage: r.RolloutPercentage,
		Enabled:           r.Enabled,
	}
}

func conditionFromRow(c sqlc.FeatureRuleCondition) domain.Condition {
	return domain.Condition{
		ID:        c.ID,
		RuleID:    c.RuleID,
		Attribute: c.Attribute,
		Operator:  c.Operator,
		Value:     c.Value,
	}
}

// ListEnabledRules returns every enabled rule across all flags with its
// conditions attached (two queries, grouped in memory — avoids N+1).
func (r *Repository) ListEnabledRules(ctx context.Context) ([]domain.Rule, error) {
	ruleRows, err := r.queries.ListAllRules(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature rules", logger.Error(err))
		return nil, err
	}
	condRows, err := r.queries.ListAllRuleConditions(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature rule conditions", logger.Error(err))
		return nil, err
	}
	condByRule := make(map[int64][]domain.Condition, len(ruleRows))
	for _, c := range condRows {
		condByRule[c.RuleID] = append(condByRule[c.RuleID], conditionFromRow(c))
	}
	rules := make([]domain.Rule, 0, len(ruleRows))
	for _, row := range ruleRows {
		rule := ruleFromRow(row)
		rule.Conditions = condByRule[row.ID]
		rules = append(rules, rule)
	}
	return rules, nil
}

// ListRulesByKey returns all rules (enabled or not) for one flag, with
// conditions, for the admin surface.
func (r *Repository) ListRulesByKey(ctx context.Context, key string) ([]domain.Rule, error) {
	ruleRows, err := r.queries.ListRulesByKey(ctx, key)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature rules by key", logger.Error(err))
		return nil, err
	}
	ruleIDs := make([]int64, 0, len(ruleRows))
	for _, row := range ruleRows {
		ruleIDs = append(ruleIDs, row.ID)
	}
	condByRule, err := r.listConditionsByRuleIDs(ctx, ruleIDs)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list feature rule conditions by rule IDs", logger.Error(err))
		return nil, err
	}

	rules := make([]domain.Rule, 0, len(ruleRows))
	for _, row := range ruleRows {
		rule := ruleFromRow(row)
		rule.Conditions = condByRule[row.ID]
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *Repository) listConditionsByRuleIDs(ctx context.Context, ruleIDs []int64) (map[int64][]domain.Condition, error) {
	condByRule := make(map[int64][]domain.Condition, len(ruleIDs))
	if len(ruleIDs) == 0 {
		return condByRule, nil
	}

	condRows, err := r.queries.ListConditionsByRuleIDs(ctx, ruleIDs)
	if err != nil {
		return nil, err
	}
	for _, c := range condRows {
		condByRule[c.RuleID] = append(condByRule[c.RuleID], conditionFromRow(c))
	}

	return condByRule, nil
}

// CreateRule inserts the rule and its conditions atomically. A committed rule
// with zero conditions matches every subject, so a partial write must never
// survive a failed condition insert.
func (r *Repository) CreateRule(ctx context.Context, rule domain.Rule) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	qtx := r.queries.WithTx(tx)

	res, err := qtx.CreateRule(ctx, sqlc.CreateRuleParams{
		FeatureKey:        rule.FeatureKey,
		Priority:          rule.Priority,
		Description:       rule.Description,
		Variant:           rule.Variant,
		RolloutPercentage: rule.RolloutPercentage,
		Enabled:           rule.Enabled,
	})
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to create feature rule", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := insertRuleConditions(ctx, tx, id, rule.Conditions); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to create rule conditions", logger.Error(err))
		return 0, err
	}
	return id, tx.Commit()
}

// ReplaceRuleConditions atomically swaps a rule's conditions for a new set.
func (r *Repository) ReplaceRuleConditions(ctx context.Context, ruleID int64, conditions []domain.Condition) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	qtx := r.queries.WithTx(tx)

	if err := qtx.DeleteConditionsByRule(ctx, ruleID); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to clear rule conditions", logger.Error(err))
		return err
	}
	if err := insertRuleConditions(ctx, tx, ruleID, conditions); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to insert rule conditions", logger.Error(err))
		return err
	}
	return tx.Commit()
}

type ruleConditionExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func insertRuleConditions(ctx context.Context, execer ruleConditionExecer, ruleID int64, conditions []domain.Condition) error {
	if len(conditions) == 0 {
		return nil
	}

	insert := sq.Insert("feature_rule_conditions").Columns("rule_id", "attribute", "operator", "value")
	for _, condition := range conditions {
		insert = insert.Values(ruleID, condition.Attribute, condition.Operator, condition.Value)
	}
	query, args, err := insert.ToSql()
	if err != nil {
		return err
	}
	_, err = execer.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) UpdateRule(ctx context.Context, rule domain.Rule) error {
	err := r.queries.UpdateRule(ctx, sqlc.UpdateRuleParams{
		ID:                rule.ID,
		Priority:          rule.Priority,
		Description:       rule.Description,
		Variant:           rule.Variant,
		RolloutPercentage: rule.RolloutPercentage,
		Enabled:           rule.Enabled,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update feature rule", logger.Error(err))
	}
	return err
}

func (r *Repository) DeleteRule(ctx context.Context, id int64) error {
	err := r.queries.DeleteRule(ctx, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete feature rule", logger.Error(err))
	}
	return err
}
