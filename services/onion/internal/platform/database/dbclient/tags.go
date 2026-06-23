package dbclient

import (
	"context"
	"database/sql"
)

type databaseTagKey struct{}

const (
	TagMain    = "main"
	TagLogging = "logging"
	TagReplica = "replica"
)

// WithTag annotates a context with the logical database dependency a query is
// using. New repository code should pass tagged contexts into QueryContext /
// ExecContext so logs and metrics can distinguish primary, logging, and
// replica reads without every callsite inventing its own labels.
func WithTag(ctx context.Context, tag string) context.Context {
	if tag == "" {
		return ctx
	}
	return context.WithValue(ctx, databaseTagKey{}, tag)
}

func TagFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	tag, _ := ctx.Value(databaseTagKey{}).(string)
	return tag
}

func PoolForTag(tag string) *sql.DB {
	switch tag {
	case TagLogging:
		return LoggingPool()
	case TagReplica:
		return ReplicaPool()
	default:
		return Pool()
	}
}

func ContextForTag(ctx context.Context, tag string) (context.Context, *sql.DB) {
	return WithTag(ctx, tag), PoolForTag(tag)
}

// TaggedDB wraps a database pool and stamps every query context with a DB tag.
// It satisfies the sqlc DBTX surface used by generated repositories.
type TaggedDB struct {
	tag string
	db  *sql.DB
}

func TaggedPool(tag string) *TaggedDB {
	return &TaggedDB{tag: tag, db: PoolForTag(tag)}
}

func (d *TaggedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(WithTag(ctx, d.tag), query, args...)
}

func (d *TaggedDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(WithTag(ctx, d.tag), query)
}

func (d *TaggedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(WithTag(ctx, d.tag), query, args...)
}

func (d *TaggedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(WithTag(ctx, d.tag), query, args...)
}

func (d *TaggedDB) PingContext(ctx context.Context) error {
	return d.db.PingContext(WithTag(ctx, d.tag))
}
