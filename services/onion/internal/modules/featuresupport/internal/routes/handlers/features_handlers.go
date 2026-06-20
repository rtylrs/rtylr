package handlers

import (
	"context"
	"strconv"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// FeaturesResponse is the shared shape for both the anonymous and authed reads:
// every resolved flag keyed by feature key, plus the top-level landing variant
// the frontend writes into the cookie for the Cloudflare worker to route on.
type FeaturesResponse struct {
	Features map[string]domain.Resolution `json:"features"`
	Variant  string                       `json:"variant"`
}

func emptyFeatures() FeaturesResponse {
	return FeaturesResponse{Features: map[string]domain.Resolution{}, Variant: application.DefaultVariant}
}

// -- POST /features (anonymous) ---------------------------------------------
// For the landing page: no auth. `subject` is a stable anonymous id the client
// generates once and persists (so the variant is sticky); `attributes` is any
// context the client wants to target on (country, app, utm, device, ...).

func GetFeaturesRoutes() []string { return []string{"/features"} }

type GetFeaturesHandler struct{ Service *application.Service }

type GetFeaturesRequest struct {
	Subject    string            `json:"subject"`
	Attributes map[string]string `json:"attributes"`
}

func (h *GetFeaturesHandler) Serve(ctx context.Context, req *GetFeaturesRequest) (FeaturesResponse, *errorModels.ServerError) {
	features, variant, err := h.Service.Resolve(ctx, domain.EvalContext{
		AnonymousID: req.Subject,
		Attributes:  req.Attributes,
	})
	if err != nil {
		// Fail safe: never break the landing page — default everyone to control.
		return emptyFeatures(), nil
	}
	return FeaturesResponse{Features: features, Variant: variant}, nil
}

// -- POST /features/me (authenticated) --------------------------------------
// For the signed-in apps: resolves for the verified user. Identity attributes
// (user_id, role, business_id, plan_id, platform) are derived server-side from
// the token and cannot be spoofed; the client may add non-identity attributes.

func GetMyFeaturesRoutes() []string { return []string{"/features/me"} }

type GetMyFeaturesHandler struct{ Service *application.Service }

type GetMyFeaturesRequest struct {
	Attributes map[string]string `json:"attributes"`
}

func (h *GetMyFeaturesHandler) Serve(ctx context.Context, req *GetMyFeaturesRequest) (FeaturesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return FeaturesResponse{}, errorModels.PermissionDeniedServerError()
	}
	u := device.User

	attrs := map[string]string{}
	for k, v := range req.Attributes {
		attrs[k] = v
	}
	// server-derived identity wins over any client-supplied value
	attrs["user_id"] = strconv.FormatInt(u.UserID, 10)
	attrs["business_id"] = strconv.FormatInt(u.BusinessID, 10)
	attrs["plan_id"] = strconv.FormatInt(u.PlanID, 10)
	if u.Role != "" {
		attrs["role"] = u.Role
	}
	if u.Platform != "" {
		attrs["platform"] = u.Platform
	}
	if email, err := h.Service.UserEmail(ctx, u.UserID); err == nil && email != "" {
		attrs["email"] = email
	}

	features, variant, err := h.Service.Resolve(ctx, domain.EvalContext{
		UserID:     u.UserID,
		Attributes: attrs,
	})
	if err != nil {
		return emptyFeatures(), nil
	}
	return FeaturesResponse{Features: features, Variant: variant}, nil
}
