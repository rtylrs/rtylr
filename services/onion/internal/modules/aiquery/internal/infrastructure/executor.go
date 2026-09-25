package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/aiquery/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

const queryTimeout = 15 * time.Second

type Executor struct{ pool *sql.DB }

// NewExecutor runs queries against the read-only replica pool. AI reads never
// touch the write primary (ReplicaPool falls back to it only if no replica is
// configured).
func NewExecutor() *Executor {
	return &Executor{pool: dbclient.ReplicaPool()}
}

// Run executes a prevalidated, parameterized query and returns up to MaxRows
// rows as ordered column/value maps. Bounded by a hard timeout.
//
// dynamic-query: squirrel
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
