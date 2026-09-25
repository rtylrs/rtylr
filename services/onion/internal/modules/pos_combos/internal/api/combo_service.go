package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/application"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poscombov1 "github.com/voxire/voxie/gen/go/pos/combo/v1"
)

type ComboAPI struct {
	Service *application.Service
}

func NewComboAPI(service *application.Service) *ComboAPI {
	return &ComboAPI{Service: service}
}

func (a *ComboAPI) ListCombos(ctx context.Context, _ *poscombov1.ListCombosRequest) (*poscombov1.ListCombosResponse, error) {
	businessID, err := authenticatedComboBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	items, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list combos")
	}

	payloads := make([]*poscombov1.Combo, len(items))
	for i, c := range items {
		payloads[i] = toProtoCombo(c)
	}
	return &poscombov1.ListCombosResponse{Combos: payloads}, nil
}

func (a *ComboAPI) CreateCombo(ctx context.Context, req *poscombov1.CreateComboRequest) (*poscombov1.CreateComboResponse, error) {
	businessID, err := authenticatedComboBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.GetPrice() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Price must be zero or greater")
	}

	device := authPkg.DeviceCtx(ctx)
	id, createErr := a.Service.Create(ctx, device.User.UserID, businessID, domain.ComboInput{
		Name:        name,
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		Items:       req.GetItems(),
		IsActive:    req.GetIsActive(),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create combo")
	}
	return &poscombov1.CreateComboResponse{ComboId: id, Message: "Combo created successfully"}, nil
}

func (a *ComboAPI) UpdateCombo(ctx context.Context, req *poscombov1.UpdateComboRequest) (*poscombov1.UpdateComboResponse, error) {
	businessID, err := authenticatedComboBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	name := strings.TrimSpace(body.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if body.GetPrice() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Price must be zero or greater")
	}

	if updateErr := a.Service.Update(ctx, businessID, req.GetId(), domain.ComboInput{
		Name:        name,
		Description: body.GetDescription(),
		Price:       body.GetPrice(),
		Items:       body.GetItems(),
		IsActive:    body.GetIsActive(),
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrComboNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Combo not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update combo")
	}
	return &poscombov1.UpdateComboResponse{Message: "Combo updated successfully"}, nil
}

func (a *ComboAPI) DeleteCombo(ctx context.Context, req *poscombov1.DeleteComboRequest) (*poscombov1.DeleteComboResponse, error) {
	businessID, err := authenticatedComboBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if deleteErr := a.Service.Delete(ctx, businessID, req.GetId()); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrComboNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Combo not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete combo")
	}
	return &poscombov1.DeleteComboResponse{Message: "Combo deleted successfully"}, nil
}

func authenticatedComboBusinessID(ctx context.Context) (int64, error) {
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

func toProtoCombo(c domain.Combo) *poscombov1.Combo {
	return &poscombov1.Combo{
		Id:          c.ID,
		OwnerId:     c.OwnerID,
		Name:        c.Name,
		Description: c.Description,
		Price:       c.Price,
		Items:       c.Items,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
