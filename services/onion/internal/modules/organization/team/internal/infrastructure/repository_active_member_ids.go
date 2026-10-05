package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) ListActiveMemberUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	rows, err := r.queries.ListActiveMemberUserIDsByBusinessID(ctx, sqlc.ListActiveMemberUserIDsByBusinessIDParams{
		ID:         businessID,
		BusinessID: businessID,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row)
	}
	return ids, nil
}
