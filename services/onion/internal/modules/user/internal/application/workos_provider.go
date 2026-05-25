package application

import (
	"context"
	"errors"
	"strings"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// GoogleUser is the normalized user info this service needs from any Google auth provider.
type GoogleUser struct {
	Email string
	Name  string
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

func (p *WorkOSProvider) GetAuthorizationURL() (string, error) {
	if p.clientID == "" {
		return "", errors.New("workos client id not configured")
	}
	url, err := p.client.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    p.clientID,
		RedirectURI: p.redirectURI,
		Provider:    "GoogleOAuth",
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
	return &GoogleUser{Email: resp.User.Email, Name: name}, nil
}
