package references

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos/references/internal/persistence/sqlc"
)

type Reader struct{ queries *sqlc.Queries }

func NewReader(db *sql.DB) *Reader { return &Reader{queries: sqlc.New(db)} }

// LocationExists reads the current committed reference in the requested business, including archived records.
func (r *Reader) LocationExists(ctx context.Context, businessID, id int64) (bool, error) {
	if businessID <= 0 || id <= 0 {
		return false, nil
	}
	return r.queries.LocationExists(ctx, sqlc.LocationExistsParams{BusinessID: businessID, ID: id})
}
