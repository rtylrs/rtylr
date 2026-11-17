// Package github is the GitHub App REST adapter for Flow. It signs app JWTs
// and mints installation tokens; neither leaves this package.
package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

const (
	DefaultAPIURL = "https://api.github.com"
	DefaultWebURL = "https://github.com"

	apiVersion   = "2022-11-28"
	maxBodyBytes = 4 << 20
	perPage      = 100
	// maxRepositoryPages caps a listing at 1,000 repositories.
	maxRepositoryPages = 10
)

type Config struct {
	AppID         string
	Slug          string
	PrivateKeyPEM []byte
	// ClientID and ClientSecret are the app's OAuth credentials for linking
	// member accounts. Both empty leaves linking off.
	ClientID     string
	ClientSecret string
	// APIURL and WebURL are overridable for tests only.
	APIURL     string
	WebURL     string
	HTTPClient *http.Client
	Now        func() time.Time
}

type Client struct {
	appID string
	slug  string
	key   *rsa.PrivateKey
	// clientID and clientSecret are empty when linking is off.
	clientID     string
	clientSecret string
	apiURL       string
	webURL       string
	http         *http.Client
	now          func() time.Time
}

func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.Slug) == "" {
		return nil, errors.New("github app id and slug are required")
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse github app private key: %w", err)
	}
	c := &Client{
		appID:        strings.TrimSpace(cfg.AppID),
		slug:         strings.TrimSpace(cfg.Slug),
		key:          key,
		clientID:     strings.TrimSpace(cfg.ClientID),
		clientSecret: strings.TrimSpace(cfg.ClientSecret),
		apiURL:       strings.TrimRight(cfg.APIURL, "/"),
		webURL:       strings.TrimRight(cfg.WebURL, "/"),
		http:         cfg.HTTPClient,
		now:          cfg.Now,
	}
	if c.apiURL == "" {
		c.apiURL = DefaultAPIURL
	}
	if c.webURL == "" {
		c.webURL = DefaultWebURL
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: 10 * time.Second}
	}
	if c.now == nil {
		c.now = time.Now
	}
	return c, nil
}

func (c *Client) InstallURL(state string) string {
	return c.webURL + "/apps/" + url.PathEscape(c.slug) + "/installations/new?state=" + url.QueryEscape(state)
}

// AppJWT signs the RS256 token GitHub requires for /app endpoints. iat is
// backdated a minute for clock drift; GitHub rejects exp beyond ten minutes.
func (c *Client) AppJWT() (string, error) {
	now := c.now()
	claims := jwt.RegisteredClaims{
		Issuer:    c.appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(c.key)
}

type installationPayload struct {
	ID      int64 `json:"id"`
	Account struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"account"`
	CreatedAt   time.Time  `json:"created_at"`
	SuspendedAt *time.Time `json:"suspended_at"`
}

func (c *Client) GetInstallation(ctx context.Context, installationID int64) (domain.GitHubInstallationInfo, error) {
	token, err := c.AppJWT()
	if err != nil {
		return domain.GitHubInstallationInfo{}, err
	}
	var out installationPayload
	if err := c.do(ctx, http.MethodGet, "/app/installations/"+strconv.FormatInt(installationID, 10), token, &out); err != nil {
		return domain.GitHubInstallationInfo{}, err
	}
	return domain.GitHubInstallationInfo{
		ID:           out.ID,
		AccountLogin: out.Account.Login,
		AccountType:  out.Account.Type,
		CreatedAt:    out.CreatedAt,
		SuspendedAt:  out.SuspendedAt,
	}, nil
}

func (c *Client) DeleteInstallation(ctx context.Context, installationID int64) error {
	token, err := c.AppJWT()
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete, "/app/installations/"+strconv.FormatInt(installationID, 10), token, nil)
}

// installationToken mints a short-lived token scoped to one installation.
func (c *Client) installationToken(ctx context.Context, installationID int64) (string, error) {
	appToken, err := c.AppJWT()
	if err != nil {
		return "", err
	}
	var out struct {
		Token string `json:"token"`
	}
	path := "/app/installations/" + strconv.FormatInt(installationID, 10) + "/access_tokens"
	if err := c.do(ctx, http.MethodPost, path, appToken, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", domain.ErrGitHubUnavailable
	}
	return out.Token, nil
}

func (c *Client) ListInstallationRepositories(ctx context.Context, installationID int64) ([]domain.GitHubRepository, error) {
	token, err := c.installationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	var repos []domain.GitHubRepository
	for page := 1; page <= maxRepositoryPages; page++ {
		var out struct {
			TotalCount   int `json:"total_count"`
			Repositories []struct {
				ID       int64  `json:"id"`
				FullName string `json:"full_name"`
				Private  bool   `json:"private"`
			} `json:"repositories"`
		}
		path := fmt.Sprintf("/installation/repositories?per_page=%d&page=%d", perPage, page)
		if err := c.do(ctx, http.MethodGet, path, token, &out); err != nil {
			return nil, err
		}
		for _, r := range out.Repositories {
			repos = append(repos, domain.GitHubRepository{ID: r.ID, FullName: r.FullName, Private: r.Private})
		}
		if len(out.Repositories) < perPage || len(repos) >= out.TotalCount {
			break
		}
	}
	return repos, nil
}

// do sends one request. A 404 maps to ErrGitHubInstallationNotFound because
// every path here is addressed by installation; anything else non-2xx is
// ErrGitHubUnavailable with the status for the log.
func (c *Client) do(ctx context.Context, method, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	req.Header.Set("User-Agent", "rtylr-flow")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrGitHubUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body := io.LimitReader(resp.Body, maxBodyBytes)
	if resp.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, body)
		return domain.ErrGitHubInstallationNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, body)
		return fmt.Errorf("%w: %s %s returned %d", domain.ErrGitHubUnavailable, method, strings.SplitN(path, "?", 2)[0], resp.StatusCode)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, body)
		return nil
	}
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return fmt.Errorf("%w: decode: %w", domain.ErrGitHubUnavailable, err)
	}
	return nil
}
