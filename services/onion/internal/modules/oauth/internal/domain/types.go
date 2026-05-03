package domain

import (
	"errors"
	"time"
)

var (
	ErrClientNotFound     = errors.New("oauth client not found")
	ErrInvalidRedirectURI = errors.New("invalid redirect_uri")
	ErrCodeNotFound       = errors.New("authorization code not found")
	ErrCodeExpired        = errors.New("authorization code expired")
	ErrCodeUsed           = errors.New("authorization code already used")
	ErrCodeVerifierFailed = errors.New("code verifier does not match challenge")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidScope       = errors.New("requested scope not allowed for this client")
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
	GetClient(clientID string) (*Client, error)
	SaveAuthCode(code *AuthCode) error
	GetAuthCode(code string) (*AuthCode, error)
	MarkAuthCodeUsed(code string) error
	SaveToken(token *Token) error
	GetTokenByRefresh(refreshToken string) (*Token, error)
	RevokeToken(refreshToken string) error
	GetUserPasswordHash(userID int64) (string, error)
	GetUserIDByEmail(email string) (int64, error)
	GetUserBusinessIDAndRole(userID int64) (businessID int64, role string, err error)
}
