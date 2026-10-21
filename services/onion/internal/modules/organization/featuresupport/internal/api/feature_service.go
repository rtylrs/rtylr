package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/organization/featuresupport/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubfeaturev1 "github.com/voxire/voxie/gen/go/hub/feature/v1"
)

type FeatureAPI struct {
	Service *application.Service
}

func NewFeatureAPI(service *application.Service) *FeatureAPI {
	return &FeatureAPI{Service: service}
}

func (a *FeatureAPI) ResolveFeatures(ctx context.Context, req *hubfeaturev1.ResolveFeaturesRequest) (*hubfeaturev1.ResolveFeaturesResponse, error) {
	features, variant, err := a.Service.Resolve(ctx, domain.EvalContext{
		AnonymousID: req.GetSubject(),
		Attributes:  req.GetAttributes(),
	})
	if err != nil {
		return emptyFeatures(), nil
	}
	return featureResponse(features, variant), nil
}

func (a *FeatureAPI) ResolveMyFeatures(ctx context.Context, req *hubfeaturev1.ResolveMyFeaturesRequest) (*hubfeaturev1.ResolveMyFeaturesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	u := device.User

	attrs := map[string]string{}
	for k, v := range req.GetAttributes() {
		attrs[k] = v
	}
	attrs["user_id"] = strconv.FormatInt(u.UserID, 10)
	attrs["business_id"] = strconv.FormatInt(u.BusinessID, 10)
	attrs["plan_id"] = strconv.FormatInt(u.PlanID, 10)
	if u.Platform != "" {
		attrs["platform"] = u.Platform
	}
	if email, err := a.Service.UserEmail(ctx, u.UserID); err == nil && email != "" {
		attrs["email"] = email
	}

	features, variant, err := a.Service.Resolve(ctx, domain.EvalContext{
		UserID:     u.UserID,
		Attributes: attrs,
	})
	if err != nil {
		empty := emptyFeatures()
		return &hubfeaturev1.ResolveMyFeaturesResponse{Features: empty.Features, Variant: empty.Variant}, nil
	}
	resp := featureResponse(features, variant)
	return &hubfeaturev1.ResolveMyFeaturesResponse{Features: resp.Features, Variant: resp.Variant}, nil
}

func (a *FeatureAPI) ListFlags(ctx context.Context, _ *hubfeaturev1.ListFlagsRequest) (*hubfeaturev1.ListFlagsResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	flags, err := a.Service.ListFlags(ctx)
	if err != nil {
		return nil, internalError("Failed to list feature flags")
	}
	payloads := make([]*hubfeaturev1.FeatureFlag, 0, len(flags))
	for _, f := range flags {
		payloads = append(payloads, flagPayload(f))
	}
	return &hubfeaturev1.ListFlagsResponse{Flags: payloads}, nil
}

func (a *FeatureAPI) UpsertFlag(ctx context.Context, req *hubfeaturev1.UpsertFlagRequest) (*hubfeaturev1.UpsertFlagResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetKey()) == "" {
		return nil, badRequest("Feature key is required")
	}
	if err := a.Service.UpsertFlag(ctx, domain.Flag{
		Key:               strings.TrimSpace(req.GetKey()),
		Description:       req.GetDescription(),
		Enabled:           req.GetEnabled(),
		RolloutPercentage: req.GetRolloutPercentage(),
		VariantOn:         req.GetVariantOn(),
		VariantOff:        req.GetVariantOff(),
	}); err != nil {
		return nil, internalError("Failed to save feature flag")
	}
	return &hubfeaturev1.UpsertFlagResponse{Message: "Feature flag saved"}, nil
}

func (a *FeatureAPI) DeleteFlag(ctx context.Context, req *hubfeaturev1.DeleteFlagRequest) (*hubfeaturev1.DeleteFlagResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetKey())
	if key == "" {
		return nil, badRequest("Feature key is required")
	}
	if err := a.Service.DeleteFlag(ctx, key); err != nil {
		return nil, internalError("Failed to delete feature flag")
	}
	return &hubfeaturev1.DeleteFlagResponse{Message: "Feature flag deleted"}, nil
}

func (a *FeatureAPI) ListTargets(ctx context.Context, req *hubfeaturev1.ListTargetsRequest) (*hubfeaturev1.ListTargetsResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetFeatureKey())
	if key == "" {
		return nil, badRequest("feature_key is required")
	}
	targets, err := a.Service.ListTargets(ctx, key)
	if err != nil {
		return nil, internalError("Failed to list feature targets")
	}
	payloads := make([]*hubfeaturev1.FeatureTarget, 0, len(targets))
	for _, t := range targets {
		payloads = append(payloads, targetPayload(t))
	}
	return &hubfeaturev1.ListTargetsResponse{Targets: payloads}, nil
}

func (a *FeatureAPI) UpsertTarget(ctx context.Context, req *hubfeaturev1.UpsertTargetRequest) (*hubfeaturev1.UpsertTargetResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetFeatureKey())
	if key == "" {
		return nil, badRequest("feature_key is required")
	}
	if req.GetOwnerId() <= 0 {
		return nil, badRequest("owner_id is required")
	}
	if err := requireTargetScope(ctx, req.GetOwnerId()); err != nil {
		return nil, err
	}
	if err := a.Service.UpsertTarget(ctx, domain.Target{
		FeatureKey: key,
		OwnerID:    req.GetOwnerId(),
		Enabled:    req.GetEnabled(),
		Variant:    strings.TrimSpace(req.GetVariant()),
	}); err != nil {
		return nil, internalError("Failed to save feature target")
	}
	return &hubfeaturev1.UpsertTargetResponse{Message: "Feature target saved"}, nil
}

func (a *FeatureAPI) DeleteTarget(ctx context.Context, req *hubfeaturev1.DeleteTargetRequest) (*hubfeaturev1.DeleteTargetResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetFeatureKey())
	if key == "" {
		return nil, badRequest("feature_key is required")
	}
	if req.GetOwnerId() <= 0 {
		return nil, badRequest("owner_id is required")
	}
	if err := requireTargetScope(ctx, req.GetOwnerId()); err != nil {
		return nil, err
	}
	if err := a.Service.DeleteTarget(ctx, key, req.GetOwnerId()); err != nil {
		return nil, internalError("Failed to delete feature target")
	}
	return &hubfeaturev1.DeleteTargetResponse{Message: "Feature target deleted"}, nil
}

func (a *FeatureAPI) ListRules(ctx context.Context, req *hubfeaturev1.ListRulesRequest) (*hubfeaturev1.ListRulesResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetFeatureKey())
	if key == "" {
		return nil, badRequest("feature_key is required")
	}
	rules, err := a.Service.ListRules(ctx, key)
	if err != nil {
		return nil, internalError("Failed to list feature rules")
	}
	payloads := make([]*hubfeaturev1.FeatureRule, 0, len(rules))
	for _, r := range rules {
		payloads = append(payloads, rulePayload(r))
	}
	return &hubfeaturev1.ListRulesResponse{Rules: payloads}, nil
}

func (a *FeatureAPI) CreateRule(ctx context.Context, req *hubfeaturev1.CreateRuleRequest) (*hubfeaturev1.CreateRuleResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	key := strings.TrimSpace(req.GetFeatureKey())
	if key == "" {
		return nil, badRequest("feature_key is required")
	}
	id, err := a.Service.CreateRule(ctx, domain.Rule{
		FeatureKey:        key,
		Priority:          req.GetPriority(),
		Description:       req.GetDescription(),
		Variant:           req.GetVariant(),
		RolloutPercentage: req.GetRolloutPercentage(),
		Enabled:           req.GetEnabled(),
		Conditions:        toConditions(req.GetConditions()),
	})
	if err != nil {
		return nil, internalError("Failed to create feature rule")
	}
	return &hubfeaturev1.CreateRuleResponse{RuleId: id, Message: "Feature rule created"}, nil
}

func (a *FeatureAPI) UpdateRule(ctx context.Context, req *hubfeaturev1.UpdateRuleRequest) (*hubfeaturev1.UpdateRuleResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == 0 {
		return nil, badRequest("rule id is required")
	}
	if err := a.Service.UpdateRule(ctx, domain.Rule{
		ID:                req.GetId(),
		Priority:          req.GetPriority(),
		Description:       req.GetDescription(),
		Variant:           req.GetVariant(),
		RolloutPercentage: req.GetRolloutPercentage(),
		Enabled:           req.GetEnabled(),
		Conditions:        toConditions(req.GetConditions()),
	}); err != nil {
		return nil, internalError("Failed to update feature rule")
	}
	return &hubfeaturev1.UpdateRuleResponse{Message: "Feature rule updated"}, nil
}

func (a *FeatureAPI) DeleteRule(ctx context.Context, req *hubfeaturev1.DeleteRuleRequest) (*hubfeaturev1.DeleteRuleResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == 0 {
		return nil, badRequest("rule id is required")
	}
	if err := a.Service.DeleteRule(ctx, req.GetId()); err != nil {
		return nil, internalError("Failed to delete feature rule")
	}
	return &hubfeaturev1.DeleteRuleResponse{Message: "Feature rule deleted"}, nil
}

func requireAdmin(ctx context.Context) error {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return onekiterrors.PermissionDenied()
	}
	if !authPkg.IsBusinessAdmin(user) && !authPkg.IsConsoleAdmin(user) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

// requireTargetScope gates feature-target writes: console_admin may target
// any tenant; every other admin may only target their own business (#817).
func requireTargetScope(ctx context.Context, requestedOwner int64) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return onekiterrors.PermissionDenied()
	}
	if authPkg.IsConsoleAdmin(user) {
		return nil
	}
	callerBusinessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return onekiterrors.PermissionDenied()
	}
	targetBusinessID, err := tenant.ResolveBusinessID(ctx, 0, requestedOwner)
	if err != nil || targetBusinessID != callerBusinessID {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func badRequest(msg string) error {
	return onekiterrors.New(http.StatusBadRequest, msg)
}

func internalError(msg string) error {
	return onekiterrors.New(http.StatusInternalServerError, msg)
}

func emptyFeatures() *hubfeaturev1.ResolveFeaturesResponse {
	return &hubfeaturev1.ResolveFeaturesResponse{Features: map[string]*hubfeaturev1.FeatureResolution{}, Variant: application.DefaultVariant}
}

func featureResponse(features map[string]domain.Resolution, variant string) *hubfeaturev1.ResolveFeaturesResponse {
	out := make(map[string]*hubfeaturev1.FeatureResolution, len(features))
	for key, value := range features {
		out[key] = resolutionPayload(value)
	}
	return &hubfeaturev1.ResolveFeaturesResponse{Features: out, Variant: variant}
}

func resolutionPayload(r domain.Resolution) *hubfeaturev1.FeatureResolution {
	return &hubfeaturev1.FeatureResolution{
		Key:     r.Key,
		Enabled: r.Enabled,
		Variant: r.Variant,
	}
}

func flagPayload(f domain.Flag) *hubfeaturev1.FeatureFlag {
	return &hubfeaturev1.FeatureFlag{
		Key:               f.Key,
		Description:       f.Description,
		Enabled:           f.Enabled,
		RolloutPercentage: f.RolloutPercentage,
		VariantOn:         f.VariantOn,
		VariantOff:        f.VariantOff,
	}
}

func targetPayload(t domain.Target) *hubfeaturev1.FeatureTarget {
	return &hubfeaturev1.FeatureTarget{
		FeatureKey: t.FeatureKey,
		OwnerId:    t.OwnerID,
		Enabled:    t.Enabled,
		Variant:    t.Variant,
	}
}

func rulePayload(r domain.Rule) *hubfeaturev1.FeatureRule {
	conds := make([]*hubfeaturev1.FeatureCondition, 0, len(r.Conditions))
	for _, c := range r.Conditions {
		conds = append(conds, &hubfeaturev1.FeatureCondition{
			Attribute: c.Attribute,
			Operator:  c.Operator,
			Value:     c.Value,
		})
	}
	return &hubfeaturev1.FeatureRule{
		Id:                r.ID,
		FeatureKey:        r.FeatureKey,
		Priority:          r.Priority,
		Description:       r.Description,
		Variant:           r.Variant,
		RolloutPercentage: r.RolloutPercentage,
		Enabled:           r.Enabled,
		Conditions:        conds,
	}
}

func toConditions(in []*hubfeaturev1.FeatureCondition) []domain.Condition {
	out := make([]domain.Condition, 0, len(in))
	for _, c := range in {
		if strings.TrimSpace(c.GetAttribute()) == "" {
			continue
		}
		out = append(out, domain.Condition{
			Attribute: strings.TrimSpace(c.GetAttribute()),
			Operator:  strings.TrimSpace(c.GetOperator()),
			Value:     c.GetValue(),
		})
	}
	return out
}
