package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/modifiers/internal/application"
	"github.com/voxire/onion/internal/modules/pos/modifiers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posmodifierv1 "github.com/voxire/voxie/gen/go/pos/modifier/v1"
)

type ModifierAPI struct {
	Service *application.Service
}

func NewModifierAPI(service *application.Service) *ModifierAPI {
	return &ModifierAPI{Service: service}
}

func (a *ModifierAPI) ListModifiers(ctx context.Context, _ *posmodifierv1.ListModifiersRequest) (*posmodifierv1.ListModifiersResponse, error) {
	businessID, err := authenticatedModifierBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	items, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list modifiers")
	}

	payloads := make([]*posmodifierv1.Modifier, len(items))
	for i, m := range items {
		payloads[i] = toProtoModifier(m)
	}
	return &posmodifierv1.ListModifiersResponse{Modifiers: payloads}, nil
}

func (a *ModifierAPI) CreateModifier(ctx context.Context, req *posmodifierv1.CreateModifierRequest) (*posmodifierv1.CreateModifierResponse, error) {
	businessID, err := authenticatedModifierBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	device := authPkg.DeviceCtx(ctx)
	id, createErr := a.Service.Create(ctx, device.User.UserID, businessID, domain.ModifierInput{
		Name:       name,
		PriceDelta: req.GetPriceDelta(),
		GroupName:  req.GetGroupName(),
		IsRequired: req.GetIsRequired(),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create modifier")
	}
	return &posmodifierv1.CreateModifierResponse{ModifierId: id, Message: "Modifier created successfully"}, nil
}

func (a *ModifierAPI) UpdateModifier(ctx context.Context, req *posmodifierv1.UpdateModifierRequest) (*posmodifierv1.UpdateModifierResponse, error) {
	businessID, err := authenticatedModifierBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	name := strings.TrimSpace(body.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	if updateErr := a.Service.Update(ctx, businessID, req.GetId(), domain.ModifierInput{
		Name:       name,
		PriceDelta: body.GetPriceDelta(),
		GroupName:  body.GetGroupName(),
		IsRequired: body.GetIsRequired(),
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrModifierNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Modifier not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update modifier")
	}
	return &posmodifierv1.UpdateModifierResponse{Message: "Modifier updated successfully"}, nil
}

func (a *ModifierAPI) DeleteModifier(ctx context.Context, req *posmodifierv1.DeleteModifierRequest) (*posmodifierv1.DeleteModifierResponse, error) {
	businessID, err := authenticatedModifierBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if deleteErr := a.Service.Delete(ctx, businessID, req.GetId()); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrModifierNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Modifier not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete modifier")
	}
	return &posmodifierv1.DeleteModifierResponse{Message: "Modifier deleted successfully"}, nil
}

func authenticatedModifierBusinessID(ctx context.Context) (int64, error) {
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

func toProtoModifier(m domain.Modifier) *posmodifierv1.Modifier {
	return &posmodifierv1.Modifier{
		Id:         m.ID,
		OwnerId:    m.OwnerID,
		Name:       m.Name,
		PriceDelta: m.PriceDelta,
		GroupName:  m.GroupName,
		IsRequired: m.IsRequired,
		CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
