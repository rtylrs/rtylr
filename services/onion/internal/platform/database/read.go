package db

import (
	"context"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	expectedQueryFuncResults = 2
)

type DatabaseType string

const (
	MainDatabase    DatabaseType = "main"
	LoggingDatabase DatabaseType = "logging"
)

var (
	ErrNoConnection          = &Error{Code: 500, Message: "no database connection available"}
	ErrNotImplemented        = &Error{Code: 501, Message: "method not implemented"}
	ErrInvalidQueryFunction  = &Error{Code: 502, Message: "invalid query function signature"}
	ErrTypeConversionFailed  = &Error{Code: 503, Message: "failed to convert row type"}
	ErrNoResults             = &Error{Code: 404, Message: "no results found"}
	ErrNoFromRowsMethod      = &Error{Code: 504, Message: "no FromRows or FromListRows method found"}
	ErrSchemaDetectionFailed = &Error{Code: 505, Message: "failed to auto-detect database schema from calling context"}
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

type Reader[T any, R any] struct {
	pool   *pgxpool.Pool
	schema string
}

func NewReadExecutor[T, R any](dbType DatabaseType) Reader[T, R] {
	targetDB, err := detectSchemaFromContext()
	if err != nil {
		// Log the error but still return a reader - the error will be caught during query execution
		panic(err) // Or handle this more gracefully depending on your error handling strategy
	}

	var pool *pgxpool.Pool

	switch dbType {
	case LoggingDatabase:
		pool = dbclient.LoggingPool()
	default: // MainDatabase - use auto-detected database
		pool = dbclient.GetPoolByName(targetDB)
	}

	return Reader[T, R]{
		pool:   pool,
		schema: targetDB, // Now stores database name for logging
	}
}

// detectSchemaFromContext detects schema based on where this is called from.
func detectSchemaFromContext() (string, error) {
	// Get the calling function's file path to determine which module is calling
	pc, file, _, ok := runtime.Caller(2) // Skip this function and NewReadExecutor/NewWriteExecutor
	if !ok {
		return "", ErrSchemaDetectionFailed
	}

	// Get function name to understand the context better
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", ErrSchemaDetectionFailed
	}

	// Parse the file path to extract the module/context
	schema, err := extractSchemaFromPath(file)
	if err != nil {
		return "", err
	}

	return schema, nil
}

// extractSchemaFromPath determines database from the file path of the caller.
func extractSchemaFromPath(filePath string) (string, error) {
	// Convert to forward slashes for consistent parsing
	normalizedPath := filepath.ToSlash(filePath)

	// Look for patterns that indicate which database to use
	// Example: /internal/modules/tooling/ -> maps to voxire database
	// Example: /internal/modules/content/ -> maps to content database
	// Example: /internal/platform/ -> maps to internal database

	if strings.Contains(normalizedPath, "/internal/platform/") {
		return "internal", nil
	}

	if strings.Contains(normalizedPath, "/internal/platforms/") {
		platformParts := strings.Split(normalizedPath, "/internal/platforms/")
		if len(platformParts) > 1 {
			platformName := strings.Split(platformParts[1], "/")[0]

			return platformName, nil
		}
	}

	if !strings.Contains(normalizedPath, "/internal/modules/") {
		return "", ErrSchemaDetectionFailed
	}

	parts := strings.Split(normalizedPath, "/internal/modules/")
	if len(parts) <= 1 {
		return "", ErrSchemaDetectionFailed
	}

	modulePath := parts[1]
	pathParts := strings.Split(modulePath, "/")

	if len(pathParts) == 0 {
		return "", ErrSchemaDetectionFailed
	}

	firstPart := pathParts[0]

	// Map module types to databases
	switch firstPart {
	case "content":
		return "content", nil
	case "tooling", "config", "user", "memory", "translation", "ai", "core":
		return "voxire", nil
	default:
		// For any other modules, default to voxire
		return "voxire", nil
	}
}

func (r Reader[T, R]) Fetch(ctx context.Context, queryFunc interface{}, args ...any) (T, error) {
	var zero T

	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return zero, ErrNoConnection
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	rows, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		return zero, err
	}

	if len(rows) == 0 {
		return zero, ErrNoResults
	}

	rowTyped, ok := rows[0].(R)
	if !ok {
		logger.Ctx(ctx).Info("Error converting row type")

		return zero, ErrTypeConversionFailed
	}

	item := new(T)
	if fromRowsErr := callFromRowsMethod(item, rowTyped); fromRowsErr != nil {
		logger.Ctx(ctx).Info("Error calling FromRows method", logger.Error(fromRowsErr))

		return zero, fromRowsErr
	}

	return *item, nil
}

func (r Reader[T, R]) FetchMapped(ctx context.Context, queryFunc func(context.Context, ...any) ([]R, error), convertFunc func(R) (string, T, error), args ...any) (map[string]T, error) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return nil, ErrNoConnection
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	rows, err := queryFunc(ctx, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[string]T)

	for _, row := range rows {
		key, val, convertErr := convertFunc(row)
		if convertErr != nil {
			logger.Ctx(ctx).Info("Error converting row", logger.Error(convertErr))

			return nil, convertErr
		}

		result[key] = val
	}

	return result, nil
}

func (r Reader[T, R]) FetchOrdered(ctx context.Context, queryFunc interface{}, args ...any) ([]T, error) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return nil, ErrNoConnection
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	rows, err := callQueryFunc(ctx, queryFunc, args...)
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(rows))

	for _, row := range rows {
		rowTyped, ok := row.(R)
		if !ok {
			logger.Ctx(ctx).Info("Error converting row type")

			return nil, ErrTypeConversionFailed
		}

		item := new(T)
		if fromRowsErr := callFromRowsMethod(item, rowTyped); fromRowsErr != nil {
			logger.Ctx(ctx).Info("Error calling FromRows method", logger.Error(fromRowsErr))

			return nil, fromRowsErr
		}

		results = append(results, *item)
	}

	return results, nil
}

func (r Reader[T, R]) Count(ctx context.Context, queryFunc func(context.Context, ...any) (int64, error), args ...any) (int, error) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return 0, ErrNoConnection
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	count, err := queryFunc(ctx, args...)
	if err != nil {
		logger.Ctx(ctx).Info("Error executing count query", logger.Error(err))

		return 0, err
	}

	return int(count), nil
}

func (r Reader[T, R]) FetchRaw(ctx context.Context, query string, args ...any) (T, bool) {
	var zero T

	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return zero, false
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	row := r.pool.QueryRow(ctx, query, args...)

	var result T

	err := row.Scan(&result)
	if err != nil {
		logger.Ctx(ctx).Info("Error scanning row", logger.Error(err))

		return zero, false
	}

	return result, true
}

func (r Reader[T, R]) FetchMappedRaw(ctx context.Context, query string, scanFunc func(pgx.Rows) (string, T, error), args ...any) (map[string]T, bool) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return nil, false
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	result := make(map[string]T)

	for rows.Next() {
		key, val, scanErr := scanFunc(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Info("Error scanning mapped row", logger.Error(scanErr))

			return nil, false
		}

		result[key] = val
	}

	return result, true
}

func (r Reader[T, R]) FetchOrderedRaw(ctx context.Context, query string, scanFunc func(pgx.Rows) (T, error), args ...any) ([]T, bool) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return nil, false
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var results []T

	for rows.Next() {
		item, scanErr := scanFunc(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Info("Error scanning row", logger.Error(scanErr))

			return nil, false
		}

		results = append(results, item)
	}

	return results, true
}

func (r Reader[T, R]) CountRaw(ctx context.Context, query string, args ...any) (int, bool) {
	if r.pool == nil {
		logger.Ctx(ctx).Error("No database connection available")

		return 0, false
	}

	ctx, cancel := readerTimeoutContext(ctx)
	defer cancel()

	var count int

	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Info("Error scanning count", logger.Error(err))

		return 0, false
	}

	return count, true
}

func readerTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	const defaultReaderTimeoutDuration = 5 * time.Second

	return context.WithTimeout(ctx, defaultReaderTimeoutDuration)
}

func callQueryFunc(ctx context.Context, queryFunc interface{}, args ...any) ([]any, error) {
	funcValue := reflect.ValueOf(queryFunc)

	callArgs := make([]reflect.Value, 0, len(args)+1)
	callArgs = append(callArgs, reflect.ValueOf(ctx))

	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	results := funcValue.Call(callArgs)

	if len(results) != expectedQueryFuncResults {
		return nil, ErrInvalidQueryFunction
	}

	if !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}

		return nil, ErrInvalidQueryFunction
	}

	resultValue := results[0]
	if resultValue.Kind() == reflect.Slice {
		length := resultValue.Len()
		slice := make([]any, length)

		for i := 0; i < length; i++ {
			slice[i] = resultValue.Index(i).Interface()
		}

		return slice, nil
	}

	return []any{resultValue.Interface()}, nil
}

func callFromRowsMethod(item, row interface{}) error {
	itemValue := reflect.ValueOf(item)

	if method := itemValue.MethodByName("FromRows"); method.IsValid() {
		results := method.Call([]reflect.Value{reflect.ValueOf(row)})
		if len(results) > 0 && !results[0].IsNil() {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}

			return ErrTypeConversionFailed
		}

		return nil
	}

	if method := itemValue.MethodByName("FromListRows"); method.IsValid() {
		results := method.Call([]reflect.Value{reflect.ValueOf(row)})
		if len(results) > 0 && !results[0].IsNil() {
			if err, ok := results[0].Interface().(error); ok {
				return err
			}

			return ErrTypeConversionFailed
		}

		return nil
	}

	return ErrNoFromRowsMethod
}
