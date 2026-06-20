package application

import (
	"context"
	"database/sql"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/infrastructure"
)

// LandingFlagKey is the flag whose resolved variant the landing page + the
// Cloudflare worker key off of to choose which repo/page to serve.
const LandingFlagKey = "landing"

// DefaultVariant is what callers get when a flag is missing or the backend
// errors — the control / "variant a".
const DefaultVariant = "a"

type repository interface {
	GetFlag(ctx context.Context, key string) (domain.Flag, error)
	ListFlags(ctx context.Context) ([]domain.Flag, error)
	UpsertFlag(ctx context.Context, f domain.Flag) error
	DeleteFlag(ctx context.Context, key string) error
	GetUserEmail(ctx context.Context, userID int64) (string, error)
	ListTargetsByOwner(ctx context.Context, ownerID int64) ([]domain.Target, error)
	ListTargetsByKey(ctx context.Context, key string) ([]domain.Target, error)
	UpsertTarget(ctx context.Context, t domain.Target) error
	DeleteTarget(ctx context.Context, key string, ownerID int64) error
	ListEnabledRules(ctx context.Context) ([]domain.Rule, error)
	ListRulesByKey(ctx context.Context, key string) ([]domain.Rule, error)
	CreateRule(ctx context.Context, rule domain.Rule) (int64, error)
	UpdateRule(ctx context.Context, rule domain.Rule) error
	DeleteRule(ctx context.Context, id int64) error
	ReplaceRuleConditions(ctx context.Context, ruleID int64, conditions []domain.Condition) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// bucketOf maps (salt, subject) to a stable bucket in [0,100). The salt mixes
// in the flag/rule so a subject doesn't land in the same bucket everywhere.
func bucketOf(salt, subject string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(salt + ":" + subject))
	return h.Sum32() % 100
}

func subjectID(ec domain.EvalContext) string {
	if ec.UserID != 0 {
		return strconv.FormatInt(ec.UserID, 10)
	}
	return ec.AnonymousID
}

func pick(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

// inList reports whether needle is one of the comma-separated entries in list.
func inList(list, needle string) bool {
	for _, part := range strings.Split(list, ",") {
		if strings.TrimSpace(part) == needle {
			return true
		}
	}
	return false
}

func numCompare(a, b, op string) bool {
	af, err1 := strconv.ParseFloat(strings.TrimSpace(a), 64)
	bf, err2 := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if err1 != nil || err2 != nil {
		return false
	}
	switch op {
	case "gt":
		return af > bf
	case "gte":
		return af >= bf
	case "lt":
		return af < bf
	case "lte":
		return af <= bf
	}
	return false
}

// matchCondition evaluates one predicate against the context attributes.
func matchCondition(c domain.Condition, attrs map[string]string) bool {
	val, present := attrs[c.Attribute]
	switch strings.ToLower(strings.TrimSpace(c.Operator)) {
	case "exists":
		return present
	case "", "eq":
		return present && val == c.Value
	case "neq":
		// matches when the attribute is absent or differs from the value
		return !present || val != c.Value
	case "in":
		return present && inList(c.Value, val)
	case "nin":
		return !present || !inList(c.Value, val)
	case "contains":
		return present && strings.Contains(val, c.Value)
	case "prefix":
		return present && strings.HasPrefix(val, c.Value)
	case "suffix":
		return present && strings.HasSuffix(val, c.Value)
	case "gt", "gte", "lt", "lte":
		return present && numCompare(val, c.Value, strings.ToLower(strings.TrimSpace(c.Operator)))
	case "regex":
		if !present {
			return false
		}
		re, err := regexp.Compile(c.Value)
		if err != nil {
			return false
		}
		return re.MatchString(val)
	default:
		return false
	}
}

// matchRule reports whether ALL of a rule's conditions match (a rule with no
// conditions targets everyone — useful as a catch-all override).
func matchRule(rule domain.Rule, attrs map[string]string) bool {
	for _, c := range rule.Conditions {
		if !matchCondition(c, attrs) {
			return false
		}
	}
	return true
}

func (s *Service) resolveOne(f domain.Flag, rules []domain.Rule, attrs map[string]string, subject string, target *domain.Target) domain.Resolution {
	// 1. an explicit per-user target always wins
	if target != nil {
		variant := target.Variant
		if variant == "" {
			if target.Enabled {
				variant = f.VariantOn
			} else {
				variant = f.VariantOff
			}
		}
		return domain.Resolution{Key: f.Key, Enabled: target.Enabled, Variant: variant}
	}

	// 2. master switch off -> off for everyone
	if !f.Enabled {
		return domain.Resolution{Key: f.Key, Enabled: false, Variant: f.VariantOff}
	}

	// 3. rules, in priority order — first matching rule decides
	for _, rule := range rules {
		if !matchRule(rule, attrs) {
			continue
		}
		onVariant := pick(rule.Variant, f.VariantOn)
		if rule.RolloutPercentage >= 100 {
			return domain.Resolution{Key: f.Key, Enabled: true, Variant: onVariant}
		}
		if rule.RolloutPercentage > 0 && subject != "" {
			salt := f.Key + ":rule:" + strconv.FormatInt(rule.ID, 10)
			if int32(bucketOf(salt, subject)) < rule.RolloutPercentage {
				return domain.Resolution{Key: f.Key, Enabled: true, Variant: onVariant}
			}
		}
		// matched the rule but bucketed out -> the flag's control variant
		return domain.Resolution{Key: f.Key, Enabled: false, Variant: f.VariantOff}
	}

	// 4. no rule matched -> the flag's global percentage rollout
	if f.RolloutPercentage >= 100 {
		return domain.Resolution{Key: f.Key, Enabled: true, Variant: f.VariantOn}
	}
	if f.RolloutPercentage <= 0 || subject == "" {
		return domain.Resolution{Key: f.Key, Enabled: false, Variant: f.VariantOff}
	}
	if int32(bucketOf(f.Key, subject)) < f.RolloutPercentage {
		return domain.Resolution{Key: f.Key, Enabled: true, Variant: f.VariantOn}
	}
	return domain.Resolution{Key: f.Key, Enabled: false, Variant: f.VariantOff}
}

// Resolve resolves every flag for the given evaluation context and also returns
// the landing variant as a convenience for the worker. Precedence per flag:
// explicit per-user target -> matching rule -> flag-level percentage rollout.
func (s *Service) Resolve(ctx context.Context, ec domain.EvalContext) (map[string]domain.Resolution, string, error) {
	flags, err := s.repo.ListFlags(ctx)
	if err != nil {
		return nil, DefaultVariant, err
	}

	targets := map[string]domain.Target{}
	if ec.UserID != 0 {
		if ts, terr := s.repo.ListTargetsByOwner(ctx, ec.UserID); terr == nil {
			for _, t := range ts {
				targets[t.FeatureKey] = t
			}
		}
	}

	rulesByFlag := map[string][]domain.Rule{}
	if allRules, rerr := s.repo.ListEnabledRules(ctx); rerr == nil {
		for _, r := range allRules {
			rulesByFlag[r.FeatureKey] = append(rulesByFlag[r.FeatureKey], r)
		}
	}

	if ec.Attributes == nil {
		ec.Attributes = map[string]string{}
	}
	subject := subjectID(ec)

	out := make(map[string]domain.Resolution, len(flags))
	variant := DefaultVariant
	for _, f := range flags {
		var tp *domain.Target
		if t, ok := targets[f.Key]; ok {
			tCopy := t
			tp = &tCopy
		}
		res := s.resolveOne(f, rulesByFlag[f.Key], ec.Attributes, subject, tp)
		out[f.Key] = res
		if f.Key == LandingFlagKey {
			variant = res.Variant
		}
	}
	return out, variant, nil
}

// --- admin / management ---

func clampPercent(p int32) int32 {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

func (s *Service) ListFlags(ctx context.Context) ([]domain.Flag, error) {
	return s.repo.ListFlags(ctx)
}

func (s *Service) UserEmail(ctx context.Context, userID int64) (string, error) {
	if userID == 0 {
		return "", nil
	}
	return s.repo.GetUserEmail(ctx, userID)
}

func (s *Service) UpsertFlag(ctx context.Context, f domain.Flag) error {
	f.RolloutPercentage = clampPercent(f.RolloutPercentage)
	f.VariantOn = pick(f.VariantOn, "b")
	f.VariantOff = pick(f.VariantOff, DefaultVariant)
	return s.repo.UpsertFlag(ctx, f)
}

func (s *Service) DeleteFlag(ctx context.Context, key string) error {
	return s.repo.DeleteFlag(ctx, key)
}

func (s *Service) ListTargets(ctx context.Context, key string) ([]domain.Target, error) {
	return s.repo.ListTargetsByKey(ctx, key)
}

func (s *Service) UpsertTarget(ctx context.Context, t domain.Target) error {
	return s.repo.UpsertTarget(ctx, t)
}

func (s *Service) DeleteTarget(ctx context.Context, key string, ownerID int64) error {
	return s.repo.DeleteTarget(ctx, key, ownerID)
}

func (s *Service) ListRules(ctx context.Context, key string) ([]domain.Rule, error) {
	return s.repo.ListRulesByKey(ctx, key)
}

func (s *Service) CreateRule(ctx context.Context, rule domain.Rule) (int64, error) {
	rule.RolloutPercentage = clampPercent(rule.RolloutPercentage)
	rule.Variant = pick(rule.Variant, "b")
	return s.repo.CreateRule(ctx, rule)
}

func (s *Service) UpdateRule(ctx context.Context, rule domain.Rule) error {
	rule.RolloutPercentage = clampPercent(rule.RolloutPercentage)
	rule.Variant = pick(rule.Variant, "b")
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return err
	}
	// conditions are replaced wholesale when provided
	return s.repo.ReplaceRuleConditions(ctx, rule.ID, rule.Conditions)
}

func (s *Service) DeleteRule(ctx context.Context, id int64) error {
	return s.repo.DeleteRule(ctx, id)
}
