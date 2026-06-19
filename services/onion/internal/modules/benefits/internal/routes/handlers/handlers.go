package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/benefits/internal/application"
	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// benefitsCacheScope namespaces the per-owner benefit plans list cache.
// hr_benefit_plans is written only within this module (Create), so a short
// TTL plus invalidation on create keeps reads consistent.
const benefitsCacheScope = "benefits"

func invalidateBenefitsCache(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(benefitsCacheScope, businessID))
	}
}

type BenefitPlanPayload struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Provider     string    `json:"provider"`
	CostEmployer float64   `json:"cost_employer"`
	CostEmployee float64   `json:"cost_employee"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

func newPayload(b domain.BenefitPlan) BenefitPlanPayload {
	return BenefitPlanPayload{
		ID: b.ID, OwnerID: b.OwnerID, Name: b.Name, Type: b.Type,
		Provider: b.Provider, CostEmployer: b.CostEmployer,
		CostEmployee: b.CostEmployee, IsActive: b.IsActive, CreatedAt: b.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/benefits"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Plans []BenefitPlanPayload `json:"plans"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	res, cacheErr := cacher.Of[ListResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(benefitsCacheScope, businessID, "list"),
		func(ctx context.Context) (ListResponse, error) {
			return h.computeList(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	res, computeErr := h.computeList(ctx, businessID)
	if computeErr != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list benefit plans"}
	}
	return res, nil
}

func (h *ListHandler) computeList(ctx context.Context, businessID int64) (ListResponse, error) {
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, err
	}
	payloads := make([]BenefitPlanPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Plans: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/benefits"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name         string  `json:"name"`
		Type         string  `json:"type"`
		Provider     string  `json:"provider"`
		CostEmployer float64 `json:"cost_employer"`
		CostEmployee float64 `json:"cost_employee"`
		IsActive     bool    `json:"is_active"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.BenefitPlanInput{
		Name: req.Name, Type: req.Type, Provider: req.Provider,
		CostEmployer: req.CostEmployer, CostEmployee: req.CostEmployee, IsActive: req.IsActive,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create benefit plan"}
	}
	invalidateBenefitsCache(ctx, businessID)
	return CreateResponse{ID: id, Message: "Benefit plan created successfully"}, nil
}
