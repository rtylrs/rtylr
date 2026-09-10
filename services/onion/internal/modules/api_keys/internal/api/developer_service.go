package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdeveloperv1 "github.com/voxire/voxie/gen/go/hub/developer/v1"
)

type DeveloperAPI struct {
	APIKeysService  hubdeveloper.APIKeysProvider
	WebhooksService hubdeveloper.WebhooksProvider
}

func NewDeveloperAPI(apiKeysService hubdeveloper.APIKeysProvider, webhooksService hubdeveloper.WebhooksProvider) *DeveloperAPI {
	return &DeveloperAPI{
		APIKeysService:  apiKeysService,
		WebhooksService: webhooksService,
	}
}

func (a *DeveloperAPI) ListApiKeys(ctx context.Context, _ *hubdeveloperv1.ListApiKeysRequest) (*hubdeveloperv1.ListApiKeysResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	keys, err := a.APIKeysService.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list API keys")
	}

	payloads := make([]*hubdeveloperv1.ApiKey, 0, len(keys))
	for _, k := range keys {
		payloads = append(payloads, toProtoAPIKey(k))
	}

	return &hubdeveloperv1.ListApiKeysResponse{Keys: payloads}, nil
}

func (a *DeveloperAPI) CreateApiKey(ctx context.Context, req *hubdeveloperv1.CreateApiKeyRequest) (*hubdeveloperv1.CreateApiKeyResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, rawKey, err := a.APIKeysService.Create(ctx, device.User.UserID, businessID, name, req.GetScopes())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create API key")
	}

	return &hubdeveloperv1.CreateApiKeyResponse{Id: id, Key: rawKey, Message: "API key created"}, nil
}

func (a *DeveloperAPI) RevokeApiKey(ctx context.Context, req *hubdeveloperv1.RevokeApiKeyRequest) (*hubdeveloperv1.RevokeApiKeyResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Key id is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.APIKeysService.Revoke(ctx, businessID, id); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to revoke API key")
	}

	return &hubdeveloperv1.RevokeApiKeyResponse{Message: "API key revoked"}, nil
}

func (a *DeveloperAPI) ListWebhooks(ctx context.Context, _ *hubdeveloperv1.ListWebhooksRequest) (*hubdeveloperv1.ListWebhooksResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	webhooks, err := a.WebhooksService.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list webhooks")
	}

	payloads := make([]*hubdeveloperv1.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		payloads = append(payloads, toProtoWebhook(w))
	}

	return &hubdeveloperv1.ListWebhooksResponse{Webhooks: payloads}, nil
}

// ListWebhookDeliveries returns delivery attempts for the caller's business.
//
// A registration whose endpoint has started failing is otherwise invisible to
// its owner: deliveries retry, exhaust their attempts and are dead-lettered
// with nothing surfacing. Filtering on status is what makes the dead ones
// findable.
//
// Read-only. Replaying a dead delivery is a mutation with real effects on
// somebody's endpoint and is deliberately not offered here.
func (a *DeveloperAPI) ListWebhookDeliveries(ctx context.Context, req *hubdeveloperv1.ListWebhookDeliveriesRequest) (*hubdeveloperv1.ListWebhookDeliveriesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolveDeliveryPagination(req.GetPage(), req.GetLimit())

	// The filters come straight from the caller and are narrowing only. The
	// listing is scoped by businessID underneath, so a webhook_id belonging to
	// another business simply matches nothing rather than reaching it.
	deliveries, total, err := a.WebhooksService.ListDeliveries(ctx, businessID, hubdeveloper.DeliveryFilter{
		WebhookID: req.GetWebhookId(),
		Status:    req.GetStatus(),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list webhook deliveries")
	}

	payloads := make([]*hubdeveloperv1.WebhookDelivery, 0, len(deliveries))
	for _, d := range deliveries {
		payloads = append(payloads, toProtoWebhookDelivery(d))
	}

	return &hubdeveloperv1.ListWebhookDeliveriesResponse{
		Deliveries: payloads,
		Pagination: buildDeliveryPagination(page, limit, total, len(payloads)),
	}, nil
}

func toProtoWebhookDelivery(d hubdeveloper.Delivery) *hubdeveloperv1.WebhookDelivery {
	out := &hubdeveloperv1.WebhookDelivery{
		Id:          d.ID,
		WebhookId:   d.WebhookID,
		EventId:     d.EventID,
		EventType:   d.EventType,
		Status:      d.Status,
		Attempts:    int32(d.Attempts),    // #nosec G115 -- bounded by max_attempts, a TINYINT.
		MaxAttempts: int32(d.MaxAttempts), // #nosec G115 -- bounded by max_attempts, a TINYINT.
		LastError:   d.LastError,
		CreatedAt:   d.CreatedAt.UTC().Format(time.RFC3339),
	}
	// Zero rather than absent when the request never got a response, which the
	// contract's field comment records.
	out.ResponseStatus = int32(d.ResponseStatus) // #nosec G115 -- an HTTP status is three digits.
	if d.NextAttemptAt != nil {
		out.NextAttemptAt = d.NextAttemptAt.UTC().Format(time.RFC3339)
	}
	if d.DeliveredAt != nil {
		out.DeliveredAt = d.DeliveredAt.UTC().Format(time.RFC3339)
	}
	return out
}

func (a *DeveloperAPI) CreateWebhook(ctx context.Context, req *hubdeveloperv1.CreateWebhookRequest) (*hubdeveloperv1.CreateWebhookResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	if strings.TrimSpace(req.GetUrl()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "URL is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.WebhooksService.Create(ctx, device.User.UserID, businessID, req.GetUrl(), req.GetEvents())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create webhook")
	}

	return &hubdeveloperv1.CreateWebhookResponse{Id: id, Message: "Webhook created"}, nil
}

func (a *DeveloperAPI) UpdateWebhook(ctx context.Context, req *hubdeveloperv1.UpdateWebhookRequest) (*hubdeveloperv1.UpdateWebhookResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Webhook id is required")
	}

	if strings.TrimSpace(req.GetUrl()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "URL is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	webhook, err := a.WebhooksService.Update(ctx, businessID, id, req.GetUrl(), req.GetEvents(), req.GetIsActive())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update webhook")
	}

	return &hubdeveloperv1.UpdateWebhookResponse{Webhook: toProtoWebhook(webhook)}, nil
}

func (a *DeveloperAPI) DeleteWebhook(ctx context.Context, req *hubdeveloperv1.DeleteWebhookRequest) (*hubdeveloperv1.DeleteWebhookResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !hasDeveloperPermission(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Webhook id is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.WebhooksService.Delete(ctx, businessID, id); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete webhook")
	}

	return &hubdeveloperv1.DeleteWebhookResponse{Message: "Webhook deleted"}, nil
}

func hasDeveloperPermission(user *authPkg.AuthUser) bool {
	return user != nil && (authPkg.HasDashboardPermission(user, "hub:api_keys:manage") || authPkg.IsBusinessAdmin(user))
}

func toProtoAPIKey(k hubdeveloper.APIKey) *hubdeveloperv1.ApiKey {
	ak := &hubdeveloperv1.ApiKey{
		Id:        k.ID,
		Name:      k.Name,
		KeyPrefix: k.KeyPrefix,
		Scopes:    k.Scopes,
		IsActive:  k.IsActive,
		CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if k.LastUsedAt != nil {
		ak.LastUsedAt = k.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if k.ExpiresAt != nil {
		ak.ExpiresAt = k.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return ak
}

func toProtoWebhook(w hubdeveloper.Webhook) *hubdeveloperv1.Webhook {
	wh := &hubdeveloperv1.Webhook{
		Id:        w.ID,
		Url:       w.URL,
		Events:    w.Events,
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if w.LastTriggeredAt != nil {
		wh.LastTriggeredAt = w.LastTriggeredAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return wh
}
