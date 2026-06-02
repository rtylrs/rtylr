package handlers

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type ConditionPayload struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

type RulePayload struct {
	ID                int64              `json:"id"`
	FeatureKey        string             `json:"feature_key"`
	Priority          int32              `json:"priority"`
	Description       string             `json:"description"`
	Variant           string             `json:"variant"`
	RolloutPercentage int32              `json:"rollout_percentage"`
	Enabled           bool               `json:"enabled"`
	Conditions        []ConditionPayload `json:"conditions"`
}

func toConditions(in []ConditionPayload) []domain.Condition {
	out := make([]domain.Condition, 0, len(in))
	for _, c := range in {
		if strings.TrimSpace(c.Attribute) == "" {
			continue
		}
		out = append(out, domain.Condition{
			Attribute: strings.TrimSpace(c.Attribute),
			Operator:  strings.TrimSpace(c.Operator),
			Value:     c.Value,
		})
	}
	return out
}

func rulePayload(r domain.Rule) RulePayload {
	conds := make([]ConditionPayload, 0, len(r.Conditions))
	for _, c := range r.Conditions {
		conds = append(conds, ConditionPayload{Attribute: c.Attribute, Operator: c.Operator, Value: c.Value})
	}
	return RulePayload{
		ID:                r.ID,
		FeatureKey:        r.FeatureKey,
		Priority:          r.Priority,
		Description:       r.Description,
		Variant:           r.Variant,
		RolloutPercentage: r.RolloutPercentage,
		Enabled:           r.Enabled,
		Conditions:        conds,
	}
}

// -- GET /features/admin/rules?feature_key= ---------------------------------

func ListRulesRoutes() []string { return []string{"/features/admin/rules"} }

type ListRulesHandler struct{ Service *application.Service }

type ListRulesRequest struct {
	FeatureKey string `json:"feature_key"`
}

type ListRulesResponse struct {
	Rules []RulePayload `json:"rules"`
}

func (h *ListRulesHandler) Serve(ctx context.Context, req *ListRulesRequest) (ListRulesResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return ListRulesResponse{}, err
	}
	if strings.TrimSpace(req.FeatureKey) == "" {
		return ListRulesResponse{}, badRequest("feature_key is required")
	}
	rules, err := h.Service.ListRules(ctx, strings.TrimSpace(req.FeatureKey))
	if err != nil {
		return ListRulesResponse{}, internalError("Failed to list feature rules")
	}
	payloads := make([]RulePayload, 0, len(rules))
	for _, r := range rules {
		payloads = append(payloads, rulePayload(r))
	}
	return ListRulesResponse{Rules: payloads}, nil
}

// -- POST /features/admin/rules (create) ------------------------------------

func CreateRuleRoutes() []string { return []string{"/features/admin/rules"} }

type CreateRuleHandler struct{ Service *application.Service }

type CreateRuleRequest struct {
	FeatureKey        string             `json:"feature_key"`
	Priority          int32              `json:"priority"`
	Description       string             `json:"description"`
	Variant           string             `json:"variant"`
	RolloutPercentage int32              `json:"rollout_percentage"`
	Enabled           bool               `json:"enabled"`
	Conditions        []ConditionPayload `json:"conditions"`
}

type CreateRuleResponse struct {
	RuleID  int64  `json:"rule_id"`
	Message string `json:"message"`
}

func (h *CreateRuleHandler) Validate(_ context.Context, req *CreateRuleRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.FeatureKey) == "" {
		return badRequest("feature_key is required")
	}
	return nil
}

func (h *CreateRuleHandler) Serve(ctx context.Context, req *CreateRuleRequest) (CreateRuleResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return CreateRuleResponse{}, err
	}
	id, err := h.Service.CreateRule(ctx, domain.Rule{
		FeatureKey:        strings.TrimSpace(req.FeatureKey),
		Priority:          req.Priority,
		Description:       req.Description,
		Variant:           req.Variant,
		RolloutPercentage: req.RolloutPercentage,
		Enabled:           req.Enabled,
		Conditions:        toConditions(req.Conditions),
	})
	if err != nil {
		return CreateRuleResponse{}, internalError("Failed to create feature rule")
	}
	return CreateRuleResponse{RuleID: id, Message: "Feature rule created"}, nil
}

// -- PUT /features/admin/rules (update + replace conditions) ----------------

func UpdateRuleRoutes() []string { return []string{"/features/admin/rules"} }

type UpdateRuleHandler struct{ Service *application.Service }

type UpdateRuleRequest struct {
	ID                int64              `json:"id"`
	Priority          int32              `json:"priority"`
	Description       string             `json:"description"`
	Variant           string             `json:"variant"`
	RolloutPercentage int32              `json:"rollout_percentage"`
	Enabled           bool               `json:"enabled"`
	Conditions        []ConditionPayload `json:"conditions"`
}

type UpdateRuleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateRuleHandler) Validate(_ context.Context, req *UpdateRuleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return badRequest("rule id is required")
	}
	return nil
}

func (h *UpdateRuleHandler) Serve(ctx context.Context, req *UpdateRuleRequest) (UpdateRuleResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return UpdateRuleResponse{}, err
	}
	if err := h.Service.UpdateRule(ctx, domain.Rule{
		ID:                req.ID,
		Priority:          req.Priority,
		Description:       req.Description,
		Variant:           req.Variant,
		RolloutPercentage: req.RolloutPercentage,
		Enabled:           req.Enabled,
		Conditions:        toConditions(req.Conditions),
	}); err != nil {
		return UpdateRuleResponse{}, internalError("Failed to update feature rule")
	}
	return UpdateRuleResponse{Message: "Feature rule updated"}, nil
}

// -- DELETE /features/admin/rules -------------------------------------------

func DeleteRuleRoutes() []string { return []string{"/features/admin/rules"} }

type DeleteRuleHandler struct{ Service *application.Service }

type DeleteRuleRequest struct {
	ID int64 `json:"id"`
}

type DeleteRuleResponse struct {
	Message string `json:"message"`
}

func (h *DeleteRuleHandler) Validate(_ context.Context, req *DeleteRuleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return badRequest("rule id is required")
	}
	return nil
}

func (h *DeleteRuleHandler) Serve(ctx context.Context, req *DeleteRuleRequest) (DeleteRuleResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return DeleteRuleResponse{}, err
	}
	if err := h.Service.DeleteRule(ctx, req.ID); err != nil {
		return DeleteRuleResponse{}, internalError("Failed to delete feature rule")
	}
	return DeleteRuleResponse{Message: "Feature rule deleted"}, nil
}
