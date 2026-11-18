package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure/persistence/sqlc"
)

// GetUserAvatar returns the stored avatar JSON, or "" for the default.
func (r *Repository) GetUserAvatar(ctx context.Context, userID int64) (string, error) {
	stored, err := r.queries.GetUserAvatar(ctx, userID)
	if err != nil {
		return "", err
	}
	return stored.String, nil
}

// SetUserAvatar stores the avatar JSON; "" clears it back to the default.
func (r *Repository) SetUserAvatar(ctx context.Context, userID int64, stored string) error {
	return r.queries.SetUserAvatar(ctx, sqlc.SetUserAvatarParams{
		Avatar: sql.NullString{String: stored, Valid: stored != ""},
		ID:     userID,
	})
}

// ListBusinessUserAvatars returns user id to stored avatar JSON for the
// business's owner and active members who chose one.
func (r *Repository) ListBusinessUserAvatars(ctx context.Context, businessID int64) (map[int64]string, error) {
	rows, err := r.queries.ListBusinessUserAvatars(ctx, sqlc.ListBusinessUserAvatarsParams{
		ID:         businessID,
		BusinessID: businessID,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[int64]string, len(rows))
	for _, row := range rows {
		if row.Avatar.Valid {
			out[row.ID] = row.Avatar.String
		}
	}
	return out, nil
}
