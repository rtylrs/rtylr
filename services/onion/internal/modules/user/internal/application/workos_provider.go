package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// GoogleUser is the normalized user info this service needs from any Google auth provider.
type GoogleUser struct {
	Email         string
	Name          string
	EmailVerified bool
}

// WorkOSProvider handles Google OAuth via WorkOS User Management.
type WorkOSProvider struct {
	client      *usermanagement.Client
	clientID    string
	redirectURI string
}

func NewWorkOSProvider(apiKey, clientID, redirectURI string) *WorkOSProvider {
	return &WorkOSProvider{
		client:      usermanagement.NewClient(apiKey),
		clientID:    clientID,
		redirectURI: redirectURI,
	}
}

// workOSProviderID maps our short provider names to WorkOS's OAuth provider
// identifiers. Returns an error for an unsupported provider.
func workOSProviderID(provider string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "google":
		return "GoogleOAuth", nil
	case "github":
		return "GitHubOAuth", nil
	case "microsoft":
		return "MicrosoftOAuth", nil
	case "apple":
		return "AppleOAuth", nil
	default:
		return "", errors.New("unsupported social provider: " + provider)
	}
}

// GetAuthorizationURL returns the WorkOS-hosted consent URL for the given
// provider (google, github, microsoft, apple). An empty provider defaults to
// Google.
func (p *WorkOSProvider) GetAuthorizationURL(provider string) (string, error) {
	if p.clientID == "" {
		return "", errors.New("workos client id not configured")
	}
	providerID, err := workOSProviderID(provider)
	if err != nil {
		return "", err
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	url, err := p.client.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    p.clientID,
		RedirectURI: p.redirectURI,
		Provider:    providerID,
		State:       state,
	})
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (p *WorkOSProvider) AuthenticateWithCode(ctx context.Context, code string) (*GoogleUser, error) {
	if p.clientID == "" {
		return nil, errors.New("workos client id not configured")
	}
	resp, err := p.client.AuthenticateWithCode(ctx, usermanagement.AuthenticateWithCodeOpts{
		ClientID: p.clientID,
		Code:     code,
	})
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(resp.User.FirstName + " " + resp.User.LastName)
	return &GoogleUser{Email: resp.User.Email, Name: name, EmailVerified: resp.User.EmailVerified}, nil
}
