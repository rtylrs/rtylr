package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

type GoogleUser struct {
	Email         string
	Name          string
	EmailVerified bool
}

type GoogleOAuthProvider struct {
	config      oauth2.Config
	userInfoURL string
}

func NewGoogleOAuthProvider(clientID, clientSecret, redirectURI string) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		userInfoURL: googleUserInfoURL,
	}
}

func (p *GoogleOAuthProvider) GetAuthorizationURL() (string, error) {
	if p == nil || p.config.ClientID == "" || p.config.ClientSecret == "" || p.config.RedirectURL == "" {
		return "", ErrGoogleOAuthNotConfigured
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	return p.config.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", "select_account")), nil
}

func (p *GoogleOAuthProvider) AuthenticateWithCode(ctx context.Context, code string) (*GoogleUser, error) {
	if p == nil || p.config.ClientID == "" || p.config.ClientSecret == "" || p.config.RedirectURL == "" {
		return nil, ErrGoogleOAuthNotConfigured
	}
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	client := p.config.Client(ctx, token)
	response, err := client.Get(p.userInfoURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, googleHTTPError{status: response.StatusCode}
	}
	var user struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode google user info: %w", err)
	}
	if user.Subject == "" || user.Email == "" {
		return nil, errors.New("google user info is missing identity fields")
	}
	return &GoogleUser{Email: user.Email, Name: user.Name, EmailVerified: user.EmailVerified}, nil
}

type googleHTTPError struct {
	status int
}

func (e googleHTTPError) Error() string {
	return fmt.Sprintf("google identity endpoint returned HTTP %d", e.status)
}

func googleOAuthUnavailable(err error) bool {
	var providerErr *oauth2.RetrieveError
	if errors.As(err, &providerErr) && providerErr.Response != nil {
		return providerErr.Response.StatusCode == http.StatusTooManyRequests || providerErr.Response.StatusCode >= http.StatusInternalServerError
	}
	var responseErr googleHTTPError
	if errors.As(err, &responseErr) {
		return responseErr.status == http.StatusTooManyRequests || responseErr.status >= http.StatusInternalServerError
	}
	return err != nil
}
