package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/bills/internal/application"
	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateBillByIDRoutes() []string { return []string{"/finance/bills/{id}"} }

type UpdateBillByIDHandler struct{ Service *application.Service }

type UpdateBillByIDRequest struct {
	SupplierID  int64   `json:"supplier_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	DueDate     string  `json:"due_date"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

type UpdateBillByIDResponse struct {
	Message string `json:"message"`
}

func (h *UpdateBillByIDHandler) Validate(_ context.Context, req *UpdateBillByIDRequest) *errorModels.ServerError {
	if req.Amount < 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "amount must not be negative"}
	}
	return nil
}

func (h *UpdateBillByIDHandler) Serve(ctx context.Context, req *UpdateBillByIDRequest) (UpdateBillByIDResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateBillByIDResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateBillByIDResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateBillByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid bill id"}
	}

	input := domain.BillInput{
		SupplierID:  req.SupplierID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		DueDate:     req.DueDate,
		Status:      req.Status,
		Description: req.Description,
	}

	if err := h.Service.Update(ctx, businessID, id, input); err != nil {
		if errors.Is(err, domain.ErrBillNotFound) {
			return UpdateBillByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Bill not found"}
		}
		logger.Ctx(ctx).Error("failed to update bill", logger.Any("owner_id", device.User.UserID), logger.Any("bill_id", id), logger.Error(err))
		return UpdateBillByIDResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update bill"}
	}

	return UpdateBillByIDResponse{Message: "Bill updated successfully"}, nil
}
