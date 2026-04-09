package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Tables ─────────────────────────────────────────────────────────────

func ListTablesRoutes() []string { return []string{"/pos/tables"} }

type ListTablesHandler struct{ Service *application.Service }

type ListTablesRequest struct {
	OwnerID    int64  `json:"ownerId"`
	LocationID int64  `json:"locationId"`
	Status     string `json:"status"`
}

type TablePayload struct {
	ID                    int64  `json:"id"`
	OwnerID               int64  `json:"ownerId"`
	LocationID            int64  `json:"locationId"`
	Name                  string `json:"name"`
	Capacity              int    `json:"capacity"`
	Status                string `json:"status"`
	CurrentSessionOrderID int64  `json:"currentSessionOrderId"`
	QRCodeURL             string `json:"qrCodeUrl"`
}

type ListTablesResponse struct {
	Tables []TablePayload `json:"tables"`
}

func newTablePayload(t domain.Table) TablePayload {
	return TablePayload{
		ID:                    t.ID,
		OwnerID:               t.OwnerID,
		LocationID:            t.LocationID,
		Name:                  t.Name,
		Capacity:              t.Capacity,
		Status:                string(t.Status),
		CurrentSessionOrderID: t.CurrentSessionOrderID,
		QRCodeURL:             t.QRCodeURL,
	}
}

func (h *ListTablesHandler) Serve(ctx context.Context, req *ListTablesRequest) (ListTablesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	ownerID := req.OwnerID
	if ownerID == 0 && device.User != nil {
		ownerID = device.User.UserID
	}

	tables, err := h.Service.ListTables(ctx, ownerID, req.LocationID, req.Status)
	if err != nil {
		return ListTablesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list tables",
		}
	}

	payloads := make([]TablePayload, 0, len(tables))
	for _, t := range tables {
		payloads = append(payloads, newTablePayload(t))
	}

	return ListTablesResponse{Tables: payloads}, nil
}

// ─── Create Table ─────────────────────────────────────────────────────────────

func CreateTableRoutes() []string { return []string{"/pos/tables"} }

type CreateTableHandler struct{ Service *application.Service }

type CreateTableRequest struct {
	LocationID int64  `json:"locationId"`
	Name       string `json:"name"`
	Capacity   int    `json:"capacity"`
	QRCodeURL  string `json:"qrCodeUrl"`
}

type CreateTableResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateTableHandler) Validate(_ context.Context, req *CreateTableRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Table name is required",
		}
	}

	return nil
}

func (h *CreateTableHandler) Serve(ctx context.Context, req *CreateTableRequest) (CreateTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTableResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	id, err := h.Service.CreateTable(ctx, ownerID, domain.TableInput{
		LocationID: req.LocationID,
		Name:       req.Name,
		Capacity:   req.Capacity,
		QRCodeURL:  req.QRCodeURL,
	})
	if err != nil {
		return CreateTableResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create table",
		}
	}

	return CreateTableResponse{ID: id, Message: "Table created"}, nil
}

// ─── Update Table ─────────────────────────────────────────────────────────────

func UpdateTableRoutes() []string { return []string{"/pos/tables"} }

type UpdateTableHandler struct{ Service *application.Service }

type UpdateTableRequest struct {
	ID         int64  `json:"id"`
	LocationID int64  `json:"locationId"`
	Name       string `json:"name"`
	Capacity   int    `json:"capacity"`
	QRCodeURL  string `json:"qrCodeUrl"`
}

type UpdateTableResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTableHandler) Validate(_ context.Context, req *UpdateTableRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Table id is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Table name is required",
		}
	}

	return nil
}

func (h *UpdateTableHandler) Serve(ctx context.Context, req *UpdateTableRequest) (UpdateTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTableResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.UpdateTable(ctx, ownerID, domain.TableInput{
		ID:         req.ID,
		LocationID: req.LocationID,
		Name:       req.Name,
		Capacity:   req.Capacity,
		QRCodeURL:  req.QRCodeURL,
	}); err != nil {
		return UpdateTableResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update table",
		}
	}

	return UpdateTableResponse{Message: "Table updated"}, nil
}

// ─── Update Table Status ──────────────────────────────────────────────────────

func UpdateTableStatusRoutes() []string { return []string{"/pos/tables/status"} }

type UpdateTableStatusHandler struct{ Service *application.Service }

type UpdateTableStatusRequest struct {
	TableID int64  `json:"tableId"`
	Status  string `json:"status"`
}

type UpdateTableStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTableStatusHandler) Validate(_ context.Context, req *UpdateTableStatusRequest) *errorModels.ServerError {
	if req.TableID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Table id is required",
		}
	}

	if req.Status == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Status is required",
		}
	}

	return nil
}

func (h *UpdateTableStatusHandler) Serve(ctx context.Context, req *UpdateTableStatusRequest) (UpdateTableStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTableStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.UpdateTableStatus(ctx, ownerID, req.TableID, domain.TableStatus(req.Status)); err != nil {
		return UpdateTableStatusResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update table status",
		}
	}

	return UpdateTableStatusResponse{Message: "Table status updated"}, nil
}

// ─── Delete Table ─────────────────────────────────────────────────────────────

func DeleteTableRoutes() []string { return []string{"/pos/tables"} }

type DeleteTableHandler struct{ Service *application.Service }

type DeleteTableRequest struct {
	TableID int64 `json:"tableId"`
}

type DeleteTableResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTableHandler) Validate(_ context.Context, req *DeleteTableRequest) *errorModels.ServerError {
	if req.TableID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Table id is required",
		}
	}

	return nil
}

func (h *DeleteTableHandler) Serve(ctx context.Context, req *DeleteTableRequest) (DeleteTableResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTableResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := device.User.UserID

	if err := h.Service.DeleteTable(ctx, ownerID, req.TableID); err != nil {
		return DeleteTableResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete table",
		}
	}

	return DeleteTableResponse{Message: "Table deleted"}, nil
}
