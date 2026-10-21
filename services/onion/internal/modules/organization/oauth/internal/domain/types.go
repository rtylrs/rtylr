package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrClientNotFound               = errors.New("oauth client not found")
	ErrInvalidRedirectURI           = errors.New("invalid redirect_uri")
	ErrCodeNotFound                 = errors.New("authorization code not found")
	ErrCodeExpired                  = errors.New("authorization code expired")
	ErrCodeUsed                     = errors.New("authorization code already used")
	ErrCodeVerifierFailed           = errors.New("code verifier does not match challenge")
	ErrPKCERequired                 = errors.New("S256 PKCE is required")
	ErrClientAuthenticationRequired = errors.New("client authentication is required")
	ErrUserNotFound                 = errors.New("oauth user identity not found")
	ErrTokenNotFound                = errors.New("token not found")
	ErrTokenRevoked                 = errors.New("token has been revoked")
	ErrTokenExpired                 = errors.New("token has expired")
	ErrInvalidScope                 = errors.New("requested scope not allowed for this client")
)

type Client struct {
	ID           string
	Name         string
	RedirectURIs []string
	Scopes       []string
	IsPublic     bool
	CreatedAt    time.Time
}

type AuthCode struct {
	Code                string
	ClientID            string
	UserID              int64
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	Used                bool
	CreatedAt           time.Time
}

type Token struct {
	ID           int64
	AccessToken  string
	RefreshToken string
	ClientID     string
	UserID       int64
	Scope        string
	ExpiresAt    time.Time
	Revoked      bool
	CreatedAt    time.Time
}

type Repository interface {
	GetClient(ctx context.Context, clientID string) (*Client, error)
	SaveAuthCode(ctx context.Context, code *AuthCode) error
	GetAuthCode(ctx context.Context, code string) (*AuthCode, error)
	MarkAuthCodeUsed(ctx context.Context, code string) error
	// ConsumeAuthCode atomically marks a code used and reports ErrCodeUsed
	// when the code was already consumed or does not exist.
	ConsumeAuthCode(ctx context.Context, code string) error
	SaveToken(ctx context.Context, token *Token) error
	GetTokenByRefresh(ctx context.Context, refreshToken string) (*Token, error)
	RevokeToken(ctx context.Context, refreshToken string) error
	// ConsumeRefreshToken atomically revokes a refresh token and reports
	// ErrTokenRevoked when it was already revoked by another request.
	ConsumeRefreshToken(ctx context.Context, refreshToken string) error
	GetUserPasswordHash(ctx context.Context, userID int64) (string, error)
	GetUserIDByEmail(ctx context.Context, email string) (int64, error)
	GetUserBusinessID(ctx context.Context, userID int64) (businessID int64, err error)
	// GetUserTokenEpoch is the revocation anchor access tokens must carry.
	GetUserTokenEpoch(ctx context.Context, userID int64) (int64, error)
}
