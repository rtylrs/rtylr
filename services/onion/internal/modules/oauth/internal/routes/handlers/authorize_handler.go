package handlers

import (
	"context"
	"net/http"
	"net/url"
	"os"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
)

// AuthorizeHandler handles GET /oauth/authorize.
// It validates the request and redirects the browser to auth.rtylr.com/oauth
// where the user completes login + consent via the frontend app.
type AuthorizeHandler struct {
	Service *application.Service
}

func (h *AuthorizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	scope := q.Get("scope")

	if _, err := h.Service.ValidateAuthRequest(r.Context(), clientID, redirectURI, scope); err != nil {
		http.Error(w, "Invalid authorization request: "+err.Error(), http.StatusBadRequest)
		return
	}

	authFrontend := os.Getenv("AUTH_APP_URL")
	if authFrontend == "" {
		authFrontend = "https://auth.rtylr.com"
	}

	dest := authFrontend + "/oauth"
	u, err := url.Parse(dest)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

// IssueRoutes returns paths for the issue endpoint called by the auth frontend.
func IssueRoutes() []string { return []string{"/oauth/authorize/issue"} }

type IssueRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type IssueResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// IssueHandler is called by auth.rtylr.com/oauth after the user logs in and grants consent.
// The user must already be authenticated (Bearer token in Authorization header).
type IssueHandler struct{ Service *application.Service }

func (h *IssueHandler) Validate(_ context.Context, req *IssueRequest) *errorModels.ServerError {
	if req.ClientID == "" || req.RedirectURI == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "client_id and redirect_uri are required",
		}
	}
	return nil
}

func (h *IssueHandler) Serve(ctx context.Context, req *IssueRequest) (IssueResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return IssueResponse{}, errorModels.PermissionDeniedServerError()
	}

	if _, err := h.Service.ValidateAuthRequest(ctx, req.ClientID, req.RedirectURI, req.Scope); err != nil {
		return IssueResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "invalid request: " + err.Error(),
		}
	}

	method := req.CodeChallengeMethod
	if method == "" {
		method = "S256"
	}

	code, err := h.Service.IssueAuthCode(
		ctx,
		req.ClientID,
		device.User.UserID,
		req.RedirectURI,
		req.Scope,
		req.CodeChallenge,
		method,
	)
	if err != nil {
		logger.Ctx(ctx).Error("failed to issue auth code")
		return IssueResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "failed to issue authorization code",
		}
	}

	callbackURL, err := url.Parse(req.RedirectURI)
	if err != nil {
		return IssueResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "invalid redirect_uri",
		}
	}

	cq := callbackURL.Query()
	cq.Set("code", code)
	if req.State != "" {
		cq.Set("state", req.State)
	}
	callbackURL.RawQuery = cq.Encode()

	return IssueResponse{RedirectURL: callbackURL.String()}, nil
}
