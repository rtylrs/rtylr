package domain

import "errors"

// ErrFlagNotFound is returned when a feature flag key does not exist.
var ErrFlagNotFound = errors.New("feature flag not found")

// Flag is a feature flag definition with a percentage rollout. A subject is
// bucketed deterministically: RolloutPercentage% of subjects get VariantOn,
// the rest get VariantOff. Enabled is the master switch.
type Flag struct {
	Key               string
	Description       string
	Enabled           bool
	RolloutPercentage int32
	VariantOn         string
	VariantOff        string
}

// Target is a per-user override that forces a feature on/off (and optionally to
// a specific variant) for one user, regardless of the rollout percentage.
type Target struct {
	FeatureKey string
	OwnerID    int64
	Enabled    bool
	Variant    string // "" means: use the flag's VariantOn/VariantOff
}

// Resolution is the resolved state of one feature for one subject.
type Resolution struct {
	Key     string
	Enabled bool
	Variant string
}

// Condition is one predicate of a rule: Attribute <Operator> Value, evaluated
// against the request's evaluation context. Operator is one of:
// eq, neq, in, nin, contains, prefix, suffix, gt, gte, lt, lte, regex, exists.
type Condition struct {
	ID        int64
	RuleID    int64
	Attribute string
	Operator  string
	Value     string
}

// Rule is an ordered targeting rule attached to a flag. It matches a subject
// when ALL of its Conditions match the evaluation context. The matched
// audience is then split: RolloutPercentage% get Variant, the rest fall to the
// flag's variant_off. Lower Priority is evaluated first; first match wins.
type Rule struct {
	ID                int64
	FeatureKey        string
	Priority          int32
	Description       string
	Variant           string
	RolloutPercentage int32
	Enabled           bool
	Conditions        []Condition
}

// EvalContext is everything known about the subject at resolution time. Attributes
// is an open map (role, plan_id, business_id, platform, app, country, email,
// or any custom key) so a rule can target by anything. UserID / AnonymousID
// drive the deterministic, sticky percentage bucketing.
type EvalContext struct {
	UserID      int64
	AnonymousID string
	Attributes  map[string]string
}
