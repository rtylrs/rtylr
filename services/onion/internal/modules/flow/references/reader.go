package references

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/flow/references/internal/persistence/sqlc"
)

type Reader struct{ queries *sqlc.Queries }

func NewReader(db *sql.DB) *Reader { return &Reader{queries: sqlc.New(db)} }

// ProjectExists reads the current committed reference in the requested business, including archived records.
func (r *Reader) ProjectExists(ctx context.Context, businessID, id int64) (bool, error) {
	if businessID <= 0 || id <= 0 {
		return false, nil
	}
	return r.queries.ProjectExists(ctx, sqlc.ProjectExistsParams{BusinessID: businessID, ID: id})
}
