package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) GetActiveMemberContact(ctx context.Context, businessID, userID int64) (string, string, bool, error) {
	contact, err := r.queries.GetActiveMemberContact(ctx, sqlc.GetActiveMemberContactParams{
		BusinessID: businessID,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", false, nil
		}
		return "", "", false, err
	}
	return contact.Email, contact.FullName, true, nil
}
