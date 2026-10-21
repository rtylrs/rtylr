package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/tables/internal/application"
	"github.com/voxire/onion/internal/modules/pos/tables/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	postablev1 "github.com/voxire/voxie/gen/go/pos/table/v1"
)

type TableAPI struct {
	Service *application.Service
}

func NewTableAPI(service *application.Service) *TableAPI {
	return &TableAPI{Service: service}
}

func (a *TableAPI) ListTables(ctx context.Context, _ *postablev1.ListTablesRequest) (*postablev1.ListTablesResponse, error) {
	businessID, err := authenticatedTableBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	tables, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list tables")
	}

	payloads := make([]*postablev1.Table, len(tables))
	for i, t := range tables {
		payloads[i] = toProtoTable(t)
	}
	return &postablev1.ListTablesResponse{Tables: payloads}, nil
}

func (a *TableAPI) CreateTable(ctx context.Context, req *postablev1.CreateTableRequest) (*postablev1.CreateTableResponse, error) {
	businessID, err := authenticatedTableBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetLabel()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Label is required")
	}

	device := authPkg.DeviceCtx(ctx)
	id, createErr := a.Service.Create(ctx, domain.TableInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		LocationID: req.LocationId,
		Label:      req.GetLabel(),
		Seats:      int(req.GetSeats()),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create table")
	}
	return &postablev1.CreateTableResponse{TableId: id, Message: "Table created successfully"}, nil
}

func (a *TableAPI) UpdateTable(ctx context.Context, req *postablev1.UpdateTableRequest) (*postablev1.UpdateTableResponse, error) {
	businessID, err := authenticatedTableBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetLabel()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Label is required")
	}
	status := strings.TrimSpace(body.GetStatus())
	if status == "" {
		status = "available"
	}
	if !validTableStatus(status) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid table status")
	}

	device := authPkg.DeviceCtx(ctx)
	if updateErr := a.Service.Update(ctx, domain.TableUpdateInput{
		ID:         req.GetId(),
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		LocationID: body.LocationId,
		Label:      body.GetLabel(),
		Seats:      int(body.GetSeats()),
		Status:     status,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrTableNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Table not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update table")
	}
	return &postablev1.UpdateTableResponse{Message: "Table updated successfully"}, nil
}

func (a *TableAPI) UpdateTableStatus(ctx context.Context, req *postablev1.UpdateTableStatusRequest) (*postablev1.UpdateTableStatusResponse, error) {
	businessID, err := authenticatedTableBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	status := strings.TrimSpace(body.GetStatus())
	if status == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	if !validTableStatus(status) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid table status")
	}

	device := authPkg.DeviceCtx(ctx)
	if updateErr := a.Service.UpdateStatus(ctx, domain.TableStatusInput{
		ID:         req.GetId(),
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Status:     status,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrTableNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Table not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update table status")
	}
	return &postablev1.UpdateTableStatusResponse{Message: "Table status updated successfully"}, nil
}

func (a *TableAPI) DeleteTable(ctx context.Context, req *postablev1.DeleteTableRequest) (*postablev1.DeleteTableResponse, error) {
	businessID, err := authenticatedTableBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if deleteErr := a.Service.Delete(ctx, businessID, req.GetId()); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrTableNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Table not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete table")
	}
	return &postablev1.DeleteTableResponse{Message: "Table deleted successfully"}, nil
}

func authenticatedTableBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func validTableStatus(status string) bool {
	for _, allowed := range domain.ValidTableStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

func toProtoTable(t domain.Table) *postablev1.Table {
	return &postablev1.Table{
		Id:         t.ID,
		OwnerId:    t.OwnerID,
		LocationId: t.LocationID,
		Label:      t.Label,
		Seats:      int32(t.Seats), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Status:     t.Status,
		CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
