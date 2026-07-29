package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support_canned/internal/application"
	"github.com/voxire/onion/internal/modules/support_canned/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportcannedv1 "github.com/voxire/voxie/gen/go/support/canned/v1"
)

type CannedResponseAPI struct {
	service *application.Service
}

func NewCannedResponseAPI(service *application.Service) *CannedResponseAPI {
	return &CannedResponseAPI{service: service}
}

func (a *CannedResponseAPI) ListCannedResponses(ctx context.Context, _ *supportcannedv1.ListCannedResponsesRequest) (*supportcannedv1.ListCannedResponsesResponse, error) {
	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list canned responses")
	}

	return &supportcannedv1.ListCannedResponsesResponse{CannedResponses: newCannedResponses(items)}, nil
}

func (a *CannedResponseAPI) CreateCannedResponse(ctx context.Context, req *supportcannedv1.CreateCannedResponseRequest) (*supportcannedv1.CreateCannedResponseResponse, error) {
	if strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if strings.TrimSpace(req.GetBody()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Body is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, domain.CannedResponseInput{
		BusinessID: businessID,
		Name:       req.GetName(),
		Body:       req.GetBody(),
		Category:   req.GetCategory(),
		Shortcut:   req.GetShortcut(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create canned response")
	}

	return &supportcannedv1.CreateCannedResponseResponse{CannedResponseId: id, Message: "Canned response created"}, nil
}

func (a *CannedResponseAPI) UpdateCannedResponse(ctx context.Context, req *supportcannedv1.UpdateCannedResponseRequest) (*supportcannedv1.UpdateCannedResponseResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.Update(ctx, businessID, req.GetId(), domain.CannedResponseInput{
		Name:     req.GetName(),
		Body:     req.GetBody(),
		Category: req.GetCategory(),
		Shortcut: req.GetShortcut(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update canned response")
	}

	return &supportcannedv1.UpdateCannedResponseResponse{Message: "Canned response updated"}, nil
}

func (a *CannedResponseAPI) DeleteCannedResponse(ctx context.Context, req *supportcannedv1.DeleteCannedResponseRequest) (*supportcannedv1.DeleteCannedResponseResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete canned response")
	}

	return &supportcannedv1.DeleteCannedResponseResponse{Message: "Canned response deleted"}, nil
}

func businessIDFromContext(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return device.User.BusinessID, nil
}

func newCannedResponses(items []domain.CannedResponse) []*supportcannedv1.CannedResponse {
	out := make([]*supportcannedv1.CannedResponse, 0, len(items))
	for _, item := range items {
		out = append(out, &supportcannedv1.CannedResponse{
			Id:         item.ID,
			BusinessId: item.BusinessID,
			Name:       item.Name,
			Body:       item.Body,
			Category:   item.Category,
			Shortcut:   item.Shortcut,
			CreatedAt:  formatTime(item.CreatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
