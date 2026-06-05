package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

const queryTimeout = 15 * time.Second

// ResolveOwnerID maps the caller's business_id to the canonical tenant id
// (businesses.owner_id) that every ai_* view's owner_id column holds. Uses the
// primary pool (admin) because the read-only AI user can't read base tables.
// Falls back to the caller's own user id.
func ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) int64 {
	if businessID <= 0 {
		return fallbackUserID
	}
	var owner int64
	err := dbclient.Pool().QueryRowContext(ctx, "SELECT owner_id FROM businesses WHERE id = ?", businessID).Scan(&owner)
	if err != nil || owner <= 0 {
		return fallbackUserID
	}
	return owner
}

type Executor struct{ pool *sql.DB }

// NewExecutor runs queries against the read-only replica pool. AI reads never
// touch the write primary (ReplicaPool falls back to it only if no replica is
// configured).
func NewExecutor() *Executor {
	return &Executor{pool: dbclient.ReplicaPool()}
}

// Run executes a prevalidated, parameterized query and returns up to MaxRows
// rows as ordered column/value maps. Bounded by a hard timeout.
func (e *Executor) Run(ctx context.Context, query string, args []any) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := e.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := make([]map[string]any, 0, domain.DefaultLimit)
	for rows.Next() && len(out) < domain.MaxRows {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if scanErr := rows.Scan(ptrs...); scanErr != nil {
			return nil, scanErr
		}

		row := make(map[string]any, len(cols))
		for i, c := range cols {
			// MySQL returns text/decimal columns as []byte; surface them as
			// strings so JSON output is readable.
			if b, ok := raw[i].([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = raw[i]
			}
		}
		out = append(out, row)
	}

	return out, rows.Err()
}
