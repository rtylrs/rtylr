package infrastructure_test

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestCreateRule_RollsBackRuleWhenConditionsFail pins B19-R04: when the
// condition insert fails (here, an attribute wider than VARCHAR(120) under
// strict mode), the rule insert must roll back with it. A committed rule with
// no conditions matches every subject, so a partial write silently overrides
// the flag for everyone.
func TestCreateRule_RollsBackRuleWhenConditionsFail(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	key := "test-create-rule-rollback"
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM feature_rules WHERE feature_key = ?", key)
	})

	_, err := repo.CreateRule(ctx, domain.Rule{
		FeatureKey:        key,
		Priority:          1,
		Variant:           "b",
		RolloutPercentage: 100,
		Enabled:           true,
		Conditions: []domain.Condition{{
			Attribute: strings.Repeat("a", 121), // exceeds attribute VARCHAR(120)
			Operator:  "eq",
			Value:     "x",
		}},
	})
	if err == nil {
		t.Fatal("CreateRule succeeded with an over-long condition attribute, want error")
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feature_rules WHERE feature_key = ?", key).Scan(&count); err != nil {
		t.Fatalf("count feature_rules: %v", err)
	}
	if count != 0 {
		t.Fatalf("feature_rules rows for %q = %d, want 0 (partial write survived)", key, count)
	}
}
