package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CompleteChecklistItemRoutes() []string {
	return []string{"/hr/onboarding/checklists/items/{id}/complete"}
}

type CompleteChecklistItemHandler struct{ Service *application.Service }

type CompleteChecklistItemRequest struct {
	AttachmentURL *string `json:"-"`
}

func (r *CompleteChecklistItemRequest) UnmarshalJSON(data []byte) error {
	r.AttachmentURL = nil
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	decode := func(raw json.RawMessage) (*string, error) {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, nil
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return &value, nil
	}

	var snake, camel *string
	var err error
	if raw, ok := wire["attachment_url"]; ok {
		snake, err = decode(raw)
		if err != nil {
			return err
		}
	}
	if raw, ok := wire["attachmentUrl"]; ok {
		camel, err = decode(raw)
		if err != nil {
			return err
		}
	}
	if snake != nil && camel != nil && *snake != *camel {
		return fmt.Errorf("attachment_url and attachmentUrl must match")
	}
	if snake != nil {
		r.AttachmentURL = snake
	} else {
		r.AttachmentURL = camel
	}
	return nil
}

func (h *CompleteChecklistItemHandler) Validate(_ context.Context, req *CompleteChecklistItemRequest) error {
	if req != nil && req.AttachmentURL != nil {
		if err := domain.ValidateAttachmentURL(*req.AttachmentURL); err != nil {
			return onekiterrors.New(http.StatusBadRequest, err.Error())
		}
	}
	return nil
}

type CompleteChecklistItemResponse struct {
	Success bool `json:"success"`
}

func (h *CompleteChecklistItemHandler) Serve(ctx context.Context, req *CompleteChecklistItemRequest) (CompleteChecklistItemResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteChecklistItemResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireOnboardingManagePermission(ctx); err != nil {
		return CompleteChecklistItemResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CompleteChecklistItemResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	itemID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid item id")
	}

	var attachmentURL *string
	if req != nil {
		attachmentURL = req.AttachmentURL
	}
	err := h.Service.CompleteChecklistItem(ctx, businessID, itemID, device.User.UserID, attachmentURL)
	if err != nil {
		if errors.Is(err, domain.ErrChecklistItemNotFound) {
			return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusNotFound, "Checklist item not found")
		}
		if errors.Is(err, domain.ErrChecklistAlreadyCompleted) {
			return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusConflict, "Checklist item is already completed")
		}
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return CompleteChecklistItemResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrInvalidAttachmentURL) {
			return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		logger.Ctx(ctx).Error("failed to complete checklist item", logger.Any("owner_id", device.User.UserID), logger.Any("item_id", itemID), logger.Error(err))
		return CompleteChecklistItemResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to complete checklist item")
	}

	return CompleteChecklistItemResponse{Success: true}, nil
}
