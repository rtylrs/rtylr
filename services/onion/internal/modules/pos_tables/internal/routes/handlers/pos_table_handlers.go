package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/application"
	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	postablev1 "github.com/voxire/voxie/gen/go/pos/table/v1"
)

func toProtoTable(t domain.Table) *postablev1.Table {
	p := &postablev1.Table{
		Id:         t.ID,
		OwnerId:    t.OwnerID,
		LocationId: t.LocationID,
		Label:      t.Label,
		Seats:      int32(t.Seats),
		Status:     t.Status,
		CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TableListRoutes() []string { return []string{"/pos/tables"} }

type (
	ListTablesHandler  struct{ Service *application.Service }
	ListTablesRequest  struct{}
	ListTablesResponse struct {
		Tables []*postablev1.Table `json:"tables"`
	}
)

func (h *ListTablesHandler) Serve(ctx context.Context, _ *ListTablesRequest) (ListTablesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTablesResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTablesResponse{}, errorModels.PermissionDeniedServerError()
	}
	tables, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListTablesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list tables"}
	}
	payloads := make([]*postablev1.Table, len(tables))
	for i, t := range tables {
		payloads[i] = toProtoTable(t)
	}
	return ListTablesResponse{Tables: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TableCreateRoutes() []string { return []string{"/pos/tables"} }

type CreateTableHandler struct{ Service *application.Service }

type CreateTableRequest struct {
	LocationID *int64 `json:"location_id"`
	Label      string `json:"label"`
	Seats      int    `json:"seats"`
}

type CreateTableResponse struct {
	TableID int64  `json:"table_id"`
	Message string `json:"message"`
}

func (h *CreateTableHandler) Validate(_ context.Context, req *CreateTableRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Label) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Label is required"}
	}
	return nil
}

func (h *CreateTableHandler) Serve(ctx context.Context, req *CreateTableRequest) (CreateTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.TableInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		LocationID: req.LocationID,
		Label:      req.Label,
		Seats:      req.Seats,
	})
	if err != nil {
		return CreateTableResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create table"}
	}
	return CreateTableResponse{TableID: id, Message: "Table created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TableUpdateRoutes() []string { return []string{"/pos/tables/{id}"} }

type UpdateTableHandler struct{ Service *application.Service }

type UpdateTableRequest struct {
	LocationID *int64 `json:"location_id"`
	Label      string `json:"label"`
	Seats      int    `json:"seats"`
	Status     string `json:"status"`
}

type UpdateTableResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTableHandler) Validate(_ context.Context, req *UpdateTableRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Label) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Label is required"}
	}
	return nil
}

func (h *UpdateTableHandler) Serve(ctx context.Context, req *UpdateTableRequest) (UpdateTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateTableResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid table ID"}
	}
	status := req.Status
	if status == "" {
		status = "available"
	}
	if updateErr := h.Service.Update(ctx, domain.TableUpdateInput{
		ID:         id,
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		LocationID: req.LocationID,
		Label:      req.Label,
		Seats:      req.Seats,
		Status:     status,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrTableNotFound) {
			return UpdateTableResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Table not found"}
		}
		return UpdateTableResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update table"}
	}
	return UpdateTableResponse{Message: "Table updated successfully"}, nil
}

// ─── Update Status ────────────────────────────────────────────────────────────

func TableUpdateStatusRoutes() []string { return []string{"/pos/tables/{id}/status"} }

type UpdateTableStatusHandler struct{ Service *application.Service }

type UpdateTableStatusRequest struct {
	Status string `json:"status"`
}

type UpdateTableStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTableStatusHandler) Validate(_ context.Context, req *UpdateTableStatusRequest) *errorModels.ServerError {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Status is required"}
	}
	if !validTableStatus(status) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid table status"}
	}
	return nil
}

func (h *UpdateTableStatusHandler) Serve(ctx context.Context, req *UpdateTableStatusRequest) (UpdateTableStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTableStatusResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateTableStatusResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateTableStatusResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid table ID"}
	}
	if updateErr := h.Service.UpdateStatus(ctx, domain.TableStatusInput{
		ID:         id,
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Status:     strings.TrimSpace(req.Status),
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrTableNotFound) {
			return UpdateTableStatusResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Table not found"}
		}
		return UpdateTableStatusResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update table status"}
	}
	return UpdateTableStatusResponse{Message: "Table status updated successfully"}, nil
}

func validTableStatus(status string) bool {
	for _, allowed := range domain.ValidTableStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TableDeleteRoutes() []string { return []string{"/pos/tables/{id}"} }

type (
	DeleteTableHandler  struct{ Service *application.Service }
	DeleteTableRequest  struct{}
	DeleteTableResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteTableHandler) Serve(ctx context.Context, _ *DeleteTableRequest) (DeleteTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteTableResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteTableResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid table ID"}
	}
	if deleteErr := h.Service.Delete(ctx, businessID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrTableNotFound) {
			return DeleteTableResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Table not found"}
		}
		return DeleteTableResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete table"}
	}
	return DeleteTableResponse{Message: "Table deleted successfully"}, nil
}
