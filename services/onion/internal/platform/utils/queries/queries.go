// Package queries builds parameterized dynamic SQL via Squirrel for callers
// that cannot use sqlc (optional filters, multi-row inserts).
package queries

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/platform/logger"
)

type Builder struct {
	columns   []string
	values    []any
	queryName string
	OrderBy   string
	Limit     int
	Offset    int
}

func NewBuilder(queryName string) *Builder {
	return &Builder{
		queryName: queryName,
	}
}

func (builder *Builder) Insert(tableName string) (queryStr string, args []any) {
	query, args, err := squirrel.Insert(tableName).Columns(builder.columns...).Values(builder.values...).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return "", nil
	}

	return builder.queryWithName(query), args
}

func (builder *Builder) Select(tableName, columns string) (queryStr string, args []any) {
	query := squirrel.Select(columns).From(tableName)

	if builder.OrderBy != "" {
		query = query.OrderBy(builder.OrderBy)
	}

	if builder.Limit != 0 {
		query = query.Limit(uint64(builder.Limit)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	if builder.Offset != 0 {
		query = query.Offset(uint64(builder.Offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	for i, column := range builder.columns {
		value := builder.values[i]
		query = query.Where(squirrel.Eq{column: value})
	}

	queryString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return "", nil
	}

	return builder.queryWithName(queryString), args
}

func (builder *Builder) SetIfNotZero(ctx context.Context, column string, val any) {
	isSet := val != reflect.Zero(reflect.TypeOf(val)).Interface()
	builder.SetIf(ctx, column, val, isSet)
}

func (builder *Builder) SetIf(ctx context.Context, column string, val any, condition bool) {
	if column == "" {
		logger.Ctx(ctx).Warn("queryAppend: column is empty")

		return
	}

	if condition {
		builder.columns = append(builder.columns, column)
		builder.values = append(builder.values, val)
	}
}

func (builder *Builder) queryWithName(query string) string {
	return fmt.Sprintf("-- :%s:\n%s", builder.queryName, query)
}
