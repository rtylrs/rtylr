package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{db: dbclient.Pool()}
}

func (r *Repository) GetClient(clientID string) (*domain.Client, error) {
	row := r.db.QueryRow(
		`SELECT id, name, redirect_uris, scopes, is_public, created_at FROM oauth_clients WHERE id = ?`,
		clientID,
	)

	var c domain.Client
	var redirectURIsJSON, scopesStr string
	var createdAt time.Time

	if err := row.Scan(&c.ID, &c.Name, &redirectURIsJSON, &scopesStr, &c.IsPublic, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(redirectURIsJSON), &c.RedirectURIs); err != nil {
		c.RedirectURIs = []string{}
	}
	c.Scopes = strings.Fields(scopesStr)
	c.CreatedAt = createdAt

	return &c, nil
}

func (r *Repository) SaveAuthCode(code *domain.AuthCode) error {
	_, err := r.db.Exec(
		`INSERT INTO oauth_auth_codes (code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scope,
		code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt,
	)
	return err
}

func (r *Repository) GetAuthCode(code string) (*domain.AuthCode, error) {
	row := r.db.QueryRow(
		`SELECT code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at, used
		 FROM oauth_auth_codes WHERE code = ?`,
		code,
	)

	var c domain.AuthCode
	var used int

	if err := row.Scan(
		&c.Code, &c.ClientID, &c.UserID, &c.RedirectURI, &c.Scope,
		&c.CodeChallenge, &c.CodeChallengeMethod, &c.ExpiresAt, &used,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCodeNotFound
		}
		return nil, err
	}

	c.Used = used == 1
	return &c, nil
}

func (r *Repository) MarkAuthCodeUsed(code string) error {
	_, err := r.db.Exec(`UPDATE oauth_auth_codes SET used = 1 WHERE code = ?`, code)
	return err
}

func (r *Repository) SaveToken(token *domain.Token) error {
	_, err := r.db.Exec(
		`INSERT INTO oauth_tokens (access_token, refresh_token, client_id, user_id, scope, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		token.AccessToken, token.RefreshToken, token.ClientID, token.UserID, token.Scope, token.ExpiresAt,
	)
	return err
}

func (r *Repository) GetTokenByRefresh(refreshToken string) (*domain.Token, error) {
	row := r.db.QueryRow(
		`SELECT id, access_token, refresh_token, client_id, user_id, scope, expires_at, revoked
		 FROM oauth_tokens WHERE refresh_token = ?`,
		refreshToken,
	)

	var t domain.Token
	var revoked int

	if err := row.Scan(
		&t.ID, &t.AccessToken, &t.RefreshToken, &t.ClientID, &t.UserID,
		&t.Scope, &t.ExpiresAt, &revoked,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	t.Revoked = revoked == 1
	return &t, nil
}

func (r *Repository) RevokeToken(refreshToken string) error {
	_, err := r.db.Exec(`UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ?`, refreshToken)
	return err
}

func (r *Repository) GetUserPasswordHash(userID int64) (string, error) {
	var hash string
	err := r.db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return hash, err
}

func (r *Repository) GetUserIDByEmail(email string) (int64, error) {
	var userID int64
	err := r.db.QueryRow(`SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return userID, err
}

func (r *Repository) GetUserBusinessIDAndRole(userID int64) (int64, string, error) {
	// cross-schema read from shared users table and user_app_permissions — for JWT claims population
	var businessID sql.NullInt64
	var role string

	err := r.db.QueryRow(
		`SELECT u.business_id, COALESCE(uap.role, '')
		 FROM users u
		 LEFT JOIN user_app_permissions uap ON uap.user_id = u.id
		 WHERE u.id = ?
		 LIMIT 1`,
		userID,
	).Scan(&businessID, &role)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return businessID.Int64, role, nil
}
