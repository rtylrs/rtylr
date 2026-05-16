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

// NewRepository accepts a ctx for parity with other modules but no longer
// stashes it on the struct — every method takes ctx as the first arg so
// request cancellation / deadlines / trace ids propagate correctly.
func NewRepository(_ context.Context) *Repository {
	return &Repository{db: dbclient.Pool()}
}

func (r *Repository) GetClient(ctx context.Context, clientID string) (*domain.Client, error) {
	row := r.db.QueryRowContext(
		ctx,
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

func (r *Repository) SaveAuthCode(ctx context.Context, code *domain.AuthCode) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO oauth_auth_codes (code, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scope,
		code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt,
	)
	return err
}

func (r *Repository) GetAuthCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	row := r.db.QueryRowContext(
		ctx,
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

func (r *Repository) MarkAuthCodeUsed(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE oauth_auth_codes SET used = 1 WHERE code = ?`, code)
	return err
}

func (r *Repository) SaveToken(ctx context.Context, token *domain.Token) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO oauth_tokens (access_token, refresh_token, client_id, user_id, scope, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		token.AccessToken, token.RefreshToken, token.ClientID, token.UserID, token.Scope, token.ExpiresAt,
	)
	return err
}

func (r *Repository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*domain.Token, error) {
	row := r.db.QueryRowContext(
		ctx,
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

func (r *Repository) RevokeToken(ctx context.Context, refreshToken string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE oauth_tokens SET revoked = 1 WHERE refresh_token = ?`, refreshToken)
	return err
}

func (r *Repository) GetUserPasswordHash(ctx context.Context, userID int64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return hash, err
}

func (r *Repository) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return userID, err
}

func (r *Repository) GetUserBusinessIDAndRole(ctx context.Context, userID int64) (int64, string, error) {
	// cross-schema read from shared users table and user_app_permissions — for JWT claims population
	var businessID sql.NullInt64
	var role string

	err := r.db.QueryRowContext(
		ctx,
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
