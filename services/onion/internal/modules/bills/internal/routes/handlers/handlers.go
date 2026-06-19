package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/bills/internal/application"
	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type BillPayload struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	SupplierID  int64     `json:"supplier_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newPayload(b domain.Bill) BillPayload {
	return BillPayload{
		ID: b.ID, OwnerID: b.OwnerID, SupplierID: b.SupplierID,
		Amount: b.Amount, Currency: b.Currency, DueDate: b.DueDate,
		Status: b.Status, Description: b.Description,
		CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func ListRoutes() []string { return []string{"/finance/bills"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Bills []BillPayload `json:"bills"`
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
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list bills"}
	}
	payloads := make([]BillPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Bills: payloads}, nil
}

func CreateRoutes() []string { return []string{"/finance/bills"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		SupplierID  int64   `json:"supplier_id"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		DueDate     string  `json:"due_date"`
		Status      string  `json:"status"`
		Description string  `json:"description"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.SupplierID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "supplier_id is required"}
	}
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "amount must be greater than zero"}
	}
	if req.DueDate == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "due_date is required"}
	}
	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "due_date must be in YYYY-MM-DD format"}
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
	input := domain.BillInput{
		SupplierID: req.SupplierID, Amount: req.Amount, Currency: req.Currency,
		DueDate: req.DueDate, Status: req.Status, Description: req.Description,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create bill"}
	}
	return CreateResponse{ID: id, Message: "Bill created successfully"}, nil
}

func UpdateRoutes() []string { return []string{"/finance/bills"} }

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID          int64   `json:"id"`
		SupplierID  int64   `json:"supplier_id"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		DueDate     string  `json:"due_date"`
		Status      string  `json:"status"`
		Description string  `json:"description"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "amount must be greater than zero"}
	}
	if req.DueDate != "" {
		if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "due_date must be in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.BillInput{
		SupplierID: req.SupplierID, Amount: req.Amount, Currency: req.Currency,
		DueDate: req.DueDate, Status: req.Status, Description: req.Description,
	}
	if err := h.Service.Update(ctx, businessID, req.ID, input); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update bill"}
	}
	return UpdateResponse{Message: "Bill updated successfully"}, nil
}
