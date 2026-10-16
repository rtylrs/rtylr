package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/oauth/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/oauth/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) GetClient(ctx context.Context, clientID string) (*domain.Client, error) {
	row, err := r.queries.GetOAuthClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}

	c := &domain.Client{
		ID:        row.ID,
		Name:      row.Name,
		IsPublic:  row.IsPublic,
		CreatedAt: row.CreatedAt.Time,
	}
	if err := json.Unmarshal([]byte(row.RedirectUris), &c.RedirectURIs); err != nil {
		c.RedirectURIs = []string{}
	}
	c.Scopes = strings.Fields(row.Scopes)
	return c, nil
}

func (r *Repository) SaveAuthCode(ctx context.Context, code *domain.AuthCode) error {
	return r.queries.SaveAuthCode(ctx, sqlc.SaveAuthCodeParams{
		Code:                code.Code,
		ClientID:            code.ClientID,
		UserID:              code.UserID,
		RedirectUri:         code.RedirectURI,
		Scope:               code.Scope,
		CodeChallenge:       code.CodeChallenge,
		CodeChallengeMethod: code.CodeChallengeMethod,
		ExpiresAt:           code.ExpiresAt,
	})
}

func (r *Repository) GetAuthCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	row, err := r.queries.GetAuthCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCodeNotFound
		}
		return nil, err
	}

	return &domain.AuthCode{
		Code:                row.Code,
		ClientID:            row.ClientID,
		UserID:              row.UserID,
		RedirectURI:         row.RedirectUri,
		Scope:               row.Scope,
		CodeChallenge:       row.CodeChallenge,
		CodeChallengeMethod: row.CodeChallengeMethod,
		ExpiresAt:           row.ExpiresAt,
		Used:                row.Used,
	}, nil
}

func (r *Repository) MarkAuthCodeUsed(ctx context.Context, code string) error {
	return r.queries.MarkAuthCodeUsed(ctx, code)
}

func (r *Repository) ConsumeAuthCode(ctx context.Context, code string) error {
	res, err := r.queries.ConsumeAuthCode(ctx, code)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrCodeUsed
	}
	return nil
}

func (r *Repository) SaveToken(ctx context.Context, token *domain.Token) error {
	return r.queries.SaveToken(ctx, sqlc.SaveTokenParams{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ClientID:     token.ClientID,
		UserID:       token.UserID,
		Scope:        token.Scope,
		ExpiresAt:    token.ExpiresAt,
	})
}

func (r *Repository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*domain.Token, error) {
	row, err := r.queries.GetTokenByRefresh(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	return &domain.Token{
		ID:           row.ID,
		AccessToken:  row.AccessToken,
		RefreshToken: row.RefreshToken,
		ClientID:     row.ClientID,
		UserID:       row.UserID,
		Scope:        row.Scope,
		ExpiresAt:    row.ExpiresAt,
		Revoked:      row.Revoked,
	}, nil
}

func (r *Repository) RevokeToken(ctx context.Context, refreshToken string) error {
	return r.queries.RevokeToken(ctx, refreshToken)
}

func (r *Repository) ConsumeRefreshToken(ctx context.Context, refreshToken string) error {
	res, err := r.queries.ConsumeRefreshToken(ctx, refreshToken)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTokenRevoked
	}
	return nil
}

func (r *Repository) GetUserPasswordHash(ctx context.Context, userID int64) (string, error) {
	// cross-schema read from shared users table — password hash for resource owner password grant
	hash, err := r.queries.GetUserPasswordHash(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("user not found")
	}
	return hash, err
}

func (r *Repository) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	// cross-schema read from shared users table — resolve user id from email for password grant
	userID, err := r.queries.GetUserIDByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	}
	return userID, err
}

func (r *Repository) GetUserBusinessID(ctx context.Context, userID int64) (int64, error) {
	businessID, err := r.queries.GetUserBusinessID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	if businessID.Valid {
		return businessID.Int64, nil
	}
	return 0, nil
}
