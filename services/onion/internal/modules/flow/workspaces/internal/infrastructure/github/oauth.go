package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// LinkingEnabled reports whether the app's OAuth client credentials are set.
func (c *Client) LinkingEnabled() bool { return c.clientID != "" && c.clientSecret != "" }

// AuthorizeURL is GitHub's user authorization page for the web application
// flow. No redirect_uri is sent, so GitHub returns to the app's first Callback
// URL; PKCE (S256) ties the returned code to this request.
func (c *Client) AuthorizeURL(state, codeChallenge string) string {
	q := url.Values{}
	q.Set("client_id", c.clientID)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	return c.webURL + "/login/oauth/authorize?" + q.Encode()
}

// ExchangeCode trades the callback code for a user access token. GitHub
// answers a bad, expired or reused code with 200 and an "error" field.
func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (string, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webURL+"/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "rtylr-flow")
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := c.send(req, &out); err != nil {
		return "", err
	}
	if out.Error != "" || out.AccessToken == "" {
		return "", fmt.Errorf("%w: %s", domain.ErrGitHubAuthorizationFailed, out.Error)
	}
	return out.AccessToken, nil
}

// GetUser reads the account behind a user access token.
func (c *Client) GetUser(ctx context.Context, userToken string) (domain.GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL+"/user", nil)
	if err != nil {
		return domain.GitHubUser{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	req.Header.Set("User-Agent", "rtylr-flow")
	var out struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.send(req, &out); err != nil {
		return domain.GitHubUser{}, err
	}
	if out.ID == 0 || out.Login == "" {
		return domain.GitHubUser{}, fmt.Errorf("%w: /user returned no account", domain.ErrGitHubUnavailable)
	}
	return domain.GitHubUser{ID: out.ID, Login: out.Login, AvatarURL: out.AvatarURL}, nil
}

// send runs an OAuth or user-token request. 401 means GitHub rejected the
// credentials; any other non-2xx is ErrGitHubUnavailable.
func (c *Client) send(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrGitHubUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body := io.LimitReader(resp.Body, maxBodyBytes)
	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, body)
		return fmt.Errorf("%w: %s returned 401", domain.ErrGitHubAuthorizationFailed, req.URL.Path)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, body)
		return fmt.Errorf("%w: %s %s returned %d", domain.ErrGitHubUnavailable, req.Method, req.URL.Path, resp.StatusCode)
	}
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return fmt.Errorf("%w: decode: %w", domain.ErrGitHubUnavailable, err)
	}
	return nil
}
