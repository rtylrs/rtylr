package api

import (
	"net/http"
	"net/url"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/application"
)

// AuthorizeHandler validates the request and redirects the browser to the auth app.
type AuthorizeHandler struct {
	Service    *application.Service
	AuthAppURL string
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

	authFrontend := h.AuthAppURL
	if authFrontend == "" {
		authFrontend = "https://sso.rtylr.com"
	}

	u, err := url.Parse(authFrontend + "/oauth")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" || u.Host == "" || u.User != nil {
		http.Error(w, "invalid auth frontend configuration", http.StatusInternalServerError)
		return
	}
	u.RawQuery = q.Encode()

	w.Header().Set("Location", u.String())
	w.WriteHeader(http.StatusFound)
}
