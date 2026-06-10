package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
)

type Repository struct {
	db *sql.DB
}

// NewRepository accepts a ctx for parity with other modules but no longer
// stashes it on the struct — every method takes ctx as the first arg so
// request cancellation / deadlines / trace ids propagate correctly.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetClient(ctx context.Context, clientID string) (*domain.Client, error) {
	query, args, err := sq.
		Select("id", "name", "redirect_uris", "scopes", "is_public", "created_at").
		From("oauth_clients").
		Where(sq.Eq{"id": clientID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)

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
	query, args, err := sq.
		Insert("oauth_auth_codes").
		Columns("code", "client_id", "user_id", "redirect_uri", "scope", "code_challenge", "code_challenge_method", "expires_at").
		Values(code.Code, code.ClientID, code.UserID, code.RedirectURI, code.Scope,
			code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetAuthCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	query, args, err := sq.
		Select("code", "client_id", "user_id", "redirect_uri", "scope", "code_challenge", "code_challenge_method", "expires_at", "used").
		From("oauth_auth_codes").
		Where(sq.Eq{"code": code}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)

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
	query, args, err := sq.
		Update("oauth_auth_codes").
		Set("used", 1).
		Where(sq.Eq{"code": code}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) SaveToken(ctx context.Context, token *domain.Token) error {
	query, args, err := sq.
		Insert("oauth_tokens").
		Columns("access_token", "refresh_token", "client_id", "user_id", "scope", "expires_at").
		Values(token.AccessToken, token.RefreshToken, token.ClientID, token.UserID, token.Scope, token.ExpiresAt).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*domain.Token, error) {
	query, args, err := sq.
		Select("id", "access_token", "refresh_token", "client_id", "user_id", "scope", "expires_at", "revoked").
		From("oauth_tokens").
		Where(sq.Eq{"refresh_token": refreshToken}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)

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
	query, args, err := sq.
		Update("oauth_tokens").
		Set("revoked", 1).
		Where(sq.Eq{"refresh_token": refreshToken}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetUserPasswordHash(ctx context.Context, userID int64) (string, error) {
	// cross-schema read from shared users table — password hash for resource owner password grant
	query, args, err := sq.
		Select("password_hash").
		From("users").
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return "", err
	}

	var hash string
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return hash, err
}

func (r *Repository) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	// cross-schema read from shared users table — resolve user id from email for password grant
	query, args, err := sq.
		Select("id").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	var userID int64
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return userID, err
}

func (r *Repository) GetUserBusinessIDAndRole(ctx context.Context, userID int64) (int64, string, error) {
	// cross-schema read from shared users table and user_app_permissions — for JWT claims population
	var businessID sql.NullInt64
	var role string

	query, args, err := sq.
		Select("u.business_id", "COALESCE(uap.role, '')").
		From("users u").
		LeftJoin("user_app_permissions uap ON uap.user_id = u.id AND uap.business_id = u.business_id").
		Where(sq.Eq{"u.id": userID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, "", err
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&businessID, &role)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return businessID.Int64, role, nil
}
