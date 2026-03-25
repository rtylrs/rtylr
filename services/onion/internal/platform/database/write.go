package db

import (
	"context"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Writer[T any, R any] struct {
	pool   *pgxpool.Pool
	schema string
}

func NewWriteExecutor[T, R any](dbType DatabaseType) Writer[T, R] {
	targetDB, err := detectSchemaFromContext()
	if err != nil {
		// Panic on schema detection failure to catch misconfigurations early
		panic(err)
	}

	var pool *pgxpool.Pool

	switch dbType {
	case LoggingDatabase:
		pool = dbclient.LoggingPool()
	default: // MainDatabase - use auto-detected database
		pool = dbclient.GetPoolByName(targetDB)
	}

	return Writer[T, R]{
		pool:   pool,
		schema: targetDB, // Now stores database name for logging
	}
}

func writerTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	const defaultWriterTimeoutDuration = 3 * time.Second

	return context.WithTimeout(ctx, defaultWriterTimeoutDuration)
}

func (w Writer[T, R]) Insert(ctx context.Context, queryFunc interface{}, args ...any) (T, error) {
	var zero T

	if w.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return zero, ErrNoConnection
	}

	ctx, cancel := writerTimeoutContext(ctx)
	defer cancel()

	result, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		logger.Ctx(ctx).Warn("Error executing insert query", logger.Error(err))

		return zero, err
	}

	if len(result) == 0 {
		return zero, ErrNoResults
	}

	rowTyped, ok := result[0].(R)
	if !ok {
		logger.Ctx(ctx).Warn("Error converting row type")

		return zero, ErrTypeConversionFailed
	}

	item := new(T)
	if fromRowsErr := callFromRowsMethod(item, rowTyped); fromRowsErr != nil {
		logger.Ctx(ctx).Warn("Error calling FromRows method", logger.Error(fromRowsErr))

		return zero, fromRowsErr
	}

	return *item, nil
}

func (w Writer[T, R]) InsertReturning(ctx context.Context, queryFunc, dest interface{}, args ...any) error {
	if w.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return ErrNoConnection
	}

	ctx, cancel := writerTimeoutContext(ctx)
	defer cancel()

	result, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		logger.Ctx(ctx).Warn("Error executing insert returning query", logger.Error(err))

		return err
	}

	if len(result) == 0 {
		return ErrNoResults
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		logger.Ctx(ctx).Warn("Destination must be a pointer")

		return ErrTypeConversionFailed
	}

	destElem := destValue.Elem()
	resultValue := reflect.ValueOf(result[0])

	if resultValue.Type().AssignableTo(destElem.Type()) {
		destElem.Set(resultValue)

		return nil
	}

	logger.Ctx(ctx).Warn("Cannot assign result to destination")

	return ErrTypeConversionFailed
}

func (w Writer[T, R]) InsertBulk(ctx context.Context, queryFunc interface{}, args ...any) ([]T, error) {
	if w.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return nil, ErrNoConnection
	}

	ctx, cancel := writerTimeoutContext(ctx)
	defer cancel()

	results, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		logger.Ctx(ctx).Warn("Error executing bulk insert query", logger.Error(err))

		return nil, err
	}

	items := make([]T, 0, len(results))

	for _, result := range results {
		rowTyped, ok := result.(R)
		if !ok {
			logger.Ctx(ctx).Warn("Error converting row type")

			return nil, ErrTypeConversionFailed
		}

		item := new(T)
		if fromRowsErr := callFromRowsMethod(item, rowTyped); fromRowsErr != nil {
			logger.Ctx(ctx).Warn("Error calling FromRows method", logger.Error(fromRowsErr))

			return nil, fromRowsErr
		}

		items = append(items, *item)
	}

	return items, nil
}

func (w Writer[T, R]) Update(ctx context.Context, queryFunc interface{}, args ...any) (T, error) {
	return w.Insert(ctx, queryFunc, args...)
}

func (w Writer[T, R]) Delete(ctx context.Context, queryFunc interface{}, args ...any) error {
	return w.Exec(ctx, queryFunc, args...)
}

func (w Writer[T, R]) Exec(ctx context.Context, queryFunc interface{}, args ...any) error {
	if w.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return ErrNoConnection
	}

	ctx, cancel := writerTimeoutContext(ctx)
	defer cancel()

	_, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		logger.Ctx(ctx).Warn("Error executing query", logger.Error(err))

		return err
	}

	return nil
}
