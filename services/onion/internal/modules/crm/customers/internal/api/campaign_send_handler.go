package api

import (
	"context"
	"errors"
	"html"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CampaignAudienceRoutes() []string { return []string{"/crm/campaigns/audience"} }
func CampaignSendRoutes() []string     { return []string{"/crm/campaigns/{id}/send"} }
func UnsubscribeRoutes() []string      { return []string{"/crm/public/unsubscribe"} }

// -- Audience ----------------------------------------------------------------

type GetCampaignAudienceHandler struct {
	CRMService *application.Service
}

type GetCampaignAudienceRequest struct {
	// SegmentID limits the audience to a segment; 0 means every customer.
	SegmentID int64 `json:"segment_id"`
}

type GetCampaignAudienceResponse struct {
	Eligible int `json:"eligible"`
	Total    int `json:"total"`
}

// Serve counts who an email campaign to the segment would reach, so the
// sender sees it before sending.
func (h *GetCampaignAudienceHandler) Serve(ctx context.Context, req *GetCampaignAudienceRequest) (GetCampaignAudienceResponse, error) {
	businessID, err := campaignBusinessID(ctx)
	if err != nil {
		return GetCampaignAudienceResponse{}, err
	}
	var segmentID *int64
	if req.SegmentID > 0 {
		segmentID = &req.SegmentID
	}
	audience, err := h.CRMService.CampaignAudience(ctx, businessID, segmentID)
	if err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return GetCampaignAudienceResponse{}, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}
		return GetCampaignAudienceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to count the audience")
	}
	return GetCampaignAudienceResponse{Eligible: audience.Eligible, Total: audience.Total}, nil
}

// -- Send --------------------------------------------------------------------

type SendCampaignHandler struct {
	CRMService *application.Service
}

// SendCampaignRequest has no body; the campaign is the {id} path segment.
type SendCampaignRequest struct{}

type SendCampaignResponse struct {
	Recipients int `json:"recipients"`
}

// Serve sends an email campaign to its opted-in audience now.
func (h *SendCampaignHandler) Serve(ctx context.Context, _ *SendCampaignRequest) (SendCampaignResponse, error) {
	campaignID, parseErr := strconv.ParseInt(serverHttp.PathParam(ctx, "id"), 10, 64)
	if parseErr != nil || campaignID <= 0 {
		return SendCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign ID")
	}
	businessID, err := campaignBusinessID(ctx)
	if err != nil {
		return SendCampaignResponse{}, err
	}
	recipients, err := h.CRMService.SendEmailCampaign(ctx, businessID, campaignID)
	switch {
	case err == nil:
		return SendCampaignResponse{Recipients: recipients}, nil
	case errors.Is(err, domain.ErrCampaignNotFound):
		return SendCampaignResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
	case errors.Is(err, domain.ErrSenderAddressRequired):
		return SendCampaignResponse{}, onekiterrors.New(http.StatusUnprocessableEntity,
			"Add your business postal address in Settings > General first. Every marketing email must include it.")
	case errors.Is(err, domain.ErrCampaignIncomplete):
		return SendCampaignResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, "Add a subject and a message before sending.")
	case errors.Is(err, domain.ErrCampaignNotSendable):
		return SendCampaignResponse{}, onekiterrors.New(http.StatusConflict, "Only an unsent email campaign can be sent.")
	default:
		logger.Ctx(ctx).Error("failed to send campaign", logger.Error(err), logger.Any("campaign_id", campaignID))
		return SendCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to send the campaign")
	}
}

func campaignBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

// -- Unsubscribe (public) ----------------------------------------------------

// UnsubscribeHandler serves the link in every marketing email. GET only
// shows a confirm button, because mail scanners prefetch links and would
// otherwise unsubscribe people; the POST from that button, or from a mail
// client's one-click unsubscribe, withdraws marketing consent.
func UnsubscribeHandler(service *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("t")
		if r.Method == http.MethodPost {
			if err := service.Unsubscribe(r.Context(), token); err != nil {
				if !errors.Is(err, domain.ErrInvalidUnsubscribe) {
					logger.Ctx(r.Context()).Error("unsubscribe failed", logger.Error(err))
					writeUnsubscribePage(w, http.StatusInternalServerError, "Something went wrong",
						"We could not unsubscribe you just now. Please try the link again in a few minutes.", "")
					return
				}
				writeUnsubscribePage(w, http.StatusBadRequest, "This link is not valid",
					"The unsubscribe link is incomplete or has been changed. Use the link from the email.", "")
				return
			}
			writeUnsubscribePage(w, http.StatusOK, "You are unsubscribed",
				"You will no longer receive marketing email from this business.", "")
			return
		}
		writeUnsubscribePage(w, http.StatusOK, "Unsubscribe from marketing email?",
			"You will stop receiving marketing email from this business. Order and account messages are not affected.",
			token)
	}
}

func writeUnsubscribePage(w http.ResponseWriter, status int, title, message, confirmToken string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex")
	w.WriteHeader(status)
	form := ""
	if confirmToken != "" {
		form = `<form method="post" action="?t=` + html.EscapeString(confirmToken) + `">` +
			`<button type="submit" style="margin-top:16px;padding:10px 18px;border:0;border-radius:8px;background:#111;color:#fff;font-size:15px;cursor:pointer">Unsubscribe</button></form>`
	}
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8">` +
		`<meta name="viewport" content="width=device-width,initial-scale=1"><title>` + html.EscapeString(title) + `</title></head>` +
		`<body style="margin:0;padding:48px 24px;font-family:Helvetica,Arial,sans-serif;background:#f6f6f6;color:#111">` +
		`<main style="max-width:440px;margin:0 auto;background:#fff;border-radius:8px;padding:28px">` +
		`<h1 style="font-size:20px;margin:0 0 8px">` + html.EscapeString(title) + `</h1>` +
		`<p style="margin:0;line-height:1.5;color:#444">` + html.EscapeString(message) + `</p>` + form +
		`</main></body></html>`))
}
