package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/benefits/internal/application"
	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type BenefitPlanPayload struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Provider     string    `json:"provider"`
	CostEmployer float64   `json:"costEmployer"`
	CostEmployee float64   `json:"costEmployee"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

func newPayload(b domain.BenefitPlan) BenefitPlanPayload {
	return BenefitPlanPayload{
		ID: b.ID, OwnerID: b.OwnerID, Name: b.Name, Type: b.Type,
		Provider: b.Provider, CostEmployer: b.CostEmployer,
		CostEmployee: b.CostEmployee, IsActive: b.IsActive, CreatedAt: b.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/benefits"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct {
	Plans []BenefitPlanPayload `json:"plans"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list benefit plans"}
	}
	payloads := make([]BenefitPlanPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Plans: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/benefits"} }

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Provider     string  `json:"provider"`
	CostEmployer float64 `json:"costEmployer"`
	CostEmployee float64 `json:"costEmployee"`
	IsActive     bool    `json:"isActive"`
}
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
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
	input := domain.BenefitPlanInput{
		Name: req.Name, Type: req.Type, Provider: req.Provider,
		CostEmployer: req.CostEmployer, CostEmployee: req.CostEmployee, IsActive: req.IsActive,
	}
	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create benefit plan"}
	}
	return CreateResponse{ID: id, Message: "Benefit plan created successfully"}, nil
}
