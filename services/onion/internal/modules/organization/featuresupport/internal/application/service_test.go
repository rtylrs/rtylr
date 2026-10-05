package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/featuresupport/internal/domain"
)

type mockRepo struct {
	flags   []domain.Flag
	targets map[int64][]domain.Target
	rules   []domain.Rule
}

func (m *mockRepo) GetFlag(_ context.Context, key string) (domain.Flag, error) {
	for _, f := range m.flags {
		if f.Key == key {
			return f, nil
		}
	}
	return domain.Flag{}, domain.ErrFlagNotFound
}
func (m *mockRepo) ListFlags(_ context.Context) ([]domain.Flag, error) { return m.flags, nil }
func (m *mockRepo) UpsertFlag(_ context.Context, _ domain.Flag) error  { return nil }
func (m *mockRepo) DeleteFlag(_ context.Context, _ string) error       { return nil }
func (m *mockRepo) GetUserEmail(_ context.Context, _ int64) (string, error) {
	return "", nil
}

func (m *mockRepo) ListTargetsByOwner(_ context.Context, ownerID int64) ([]domain.Target, error) {
	return m.targets[ownerID], nil
}

func (m *mockRepo) ListTargetsByKey(_ context.Context, _ string) ([]domain.Target, error) {
	return nil, nil
}
func (m *mockRepo) UpsertTarget(_ context.Context, _ domain.Target) error     { return nil }
func (m *mockRepo) DeleteTarget(_ context.Context, _ string, _ int64) error   { return nil }
func (m *mockRepo) ListEnabledRules(_ context.Context) ([]domain.Rule, error) { return m.rules, nil }
func (m *mockRepo) ListRulesByKey(_ context.Context, _ string) ([]domain.Rule, error) {
	return m.rules, nil
}
func (m *mockRepo) CreateRule(_ context.Context, _ domain.Rule) (int64, error) { return 1, nil }
func (m *mockRepo) UpdateRule(_ context.Context, _ domain.Rule) error          { return nil }
func (m *mockRepo) DeleteRule(_ context.Context, _ int64) error                { return nil }
func (m *mockRepo) ReplaceRuleConditions(_ context.Context, _ int64, _ []domain.Condition) error {
	return nil
}

func landingFlag(rollout int32) domain.Flag {
	return domain.Flag{Key: LandingFlagKey, Enabled: true, RolloutPercentage: rollout, VariantOn: "b", VariantOff: "a"}
}

func TestMatchCondition(t *testing.T) {
	attrs := map[string]string{"role": "owner", "plan_id": "3", "email": "x@acme.com", "country": "LB"}
	cases := []struct {
		cond domain.Condition
		want bool
	}{
		{domain.Condition{Attribute: "role", Operator: "eq", Value: "owner"}, true},
		{domain.Condition{Attribute: "role", Operator: "eq", Value: "staff"}, false},
		{domain.Condition{Attribute: "role", Operator: "in", Value: "admin,owner"}, true},
		{domain.Condition{Attribute: "role", Operator: "nin", Value: "admin,staff"}, true},
		{domain.Condition{Attribute: "role", Operator: "neq", Value: "owner"}, false},
		{domain.Condition{Attribute: "missing", Operator: "exists", Value: ""}, false},
		{domain.Condition{Attribute: "role", Operator: "exists", Value: ""}, true},
		{domain.Condition{Attribute: "plan_id", Operator: "gte", Value: "3"}, true},
		{domain.Condition{Attribute: "plan_id", Operator: "gt", Value: "3"}, false},
		{domain.Condition{Attribute: "email", Operator: "suffix", Value: "@acme.com"}, true},
		{domain.Condition{Attribute: "email", Operator: "contains", Value: "acme"}, true},
		{domain.Condition{Attribute: "country", Operator: "regex", Value: "^(LB|US)$"}, true},
	}
	for i, c := range cases {
		if got := matchCondition(c.cond, attrs); got != c.want {
			t.Errorf("case %d (%s %s %q): got %v want %v", i, c.cond.Attribute, c.cond.Operator, c.cond.Value, got, c.want)
		}
	}
}

func TestResolve_FullRollout(t *testing.T) {
	s := &Service{repo: &mockRepo{flags: []domain.Flag{landingFlag(100)}}}
	_, variant, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "anon-1"})
	if variant != "b" {
		t.Errorf("100%% rollout should give variant b, got %q", variant)
	}
}

func TestResolve_ZeroRollout(t *testing.T) {
	s := &Service{repo: &mockRepo{flags: []domain.Flag{landingFlag(0)}}}
	_, variant, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "anon-1"})
	if variant != "a" {
		t.Errorf("0%% rollout should give control variant a, got %q", variant)
	}
}

func TestResolve_Deterministic(t *testing.T) {
	s := &Service{repo: &mockRepo{flags: []domain.Flag{landingFlag(50)}}}
	_, v1, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "stable-subject"})
	_, v2, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "stable-subject"})
	if v1 != v2 {
		t.Errorf("same subject must be sticky: %q vs %q", v1, v2)
	}
}

func TestResolve_PerUserTargetWins(t *testing.T) {
	repo := &mockRepo{
		flags:   []domain.Flag{landingFlag(0)}, // rollout 0 -> would be 'a'
		targets: map[int64][]domain.Target{7: {{FeatureKey: LandingFlagKey, OwnerID: 7, Enabled: true, Variant: "b"}}},
	}
	s := &Service{repo: repo}
	_, variant, _ := s.Resolve(context.Background(), domain.EvalContext{UserID: 7})
	if variant != "b" {
		t.Errorf("per-user target should force variant b, got %q", variant)
	}
}

func TestResolve_RuleByAttribute(t *testing.T) {
	repo := &mockRepo{
		flags: []domain.Flag{landingFlag(0)}, // default would be 'a'
		rules: []domain.Rule{{
			ID: 1, FeatureKey: LandingFlagKey, Variant: "b", RolloutPercentage: 100, Enabled: true,
			Conditions: []domain.Condition{{Attribute: "country", Operator: "eq", Value: "LB"}},
		}},
	}
	s := &Service{repo: repo}
	// matching attribute -> rule serves b
	_, vMatch, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "z", Attributes: map[string]string{"country": "LB"}})
	if vMatch != "b" {
		t.Errorf("rule should match country=LB and serve b, got %q", vMatch)
	}
	// non-matching -> falls to flag default (a)
	_, vMiss, _ := s.Resolve(context.Background(), domain.EvalContext{AnonymousID: "z", Attributes: map[string]string{"country": "US"}})
	if vMiss != "a" {
		t.Errorf("rule should not match country=US, expected control a, got %q", vMiss)
	}
}
