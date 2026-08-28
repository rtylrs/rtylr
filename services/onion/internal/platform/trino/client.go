// Package trino is an optional analytics warehouse client backed by Trino.
// When TRINO_ENABLED=false the package is inert — Connect is a no-op and
// Pool returns nil.
package trino

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	trinodriver "github.com/trinodb/trino-go-client/trino"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	customHTTPClientName = "onion"

	defaultMaxOpenConns    = 5
	defaultMaxIdleConns    = 2
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 10 * time.Minute
	defaultConnectTimeout  = 10 * time.Second
	defaultQueryTimeout    = 2 * time.Minute
)

var (
	ErrDisabled      = errors.New("trino is disabled")
	ErrNotConfigured = errors.New("trino is not configured")
	ErrNotConnected  = errors.New("trino is not connected")

	clientInstance *Client
	clientOnce     sync.Once
	httpOnce       sync.Once
	httpClientErr  error
)

// Client wraps a Trino sql.DB pool.
type Client struct {
	db *sql.DB
}

// Connect initialises the Trino pool when enabled. When Trino is disabled this
// is a no-op. When enabled but unreachable, the API still starts — callers see
// Available() == false and Query returns ErrNotConnected.
func Connect(ctx context.Context) error {
	if !config.IsTrinoConfigured() {
		if config.IsTrinoEnabled() {
			logger.Ctx(ctx).Warn("trino is enabled but missing TRINO_HOST or TRINO_USER; analytics queries disabled")
		} else {
			logger.Ctx(ctx).Info("trino analytics disabled")
		}
		return nil
	}

	c := getClient()
	if c.db != nil {
		return nil
	}

	if err := registerHTTPClient(); err != nil {
		logger.Ctx(ctx).Warn("failed to register trino HTTP client; analytics queries disabled", logger.Error(err))
		return nil
	}

	db, err := openDB(ctx)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to connect to trino; analytics queries disabled", logger.Error(err))
		return nil
	}

	c.db = db
	logger.Ctx(ctx).Info("connected to trino",
		logger.String("host", config.GetTrinoHost()),
		logger.String("port", config.GetTrinoPort()),
		logger.String("catalog", config.GetTrinoCatalog()),
		logger.String("schema", config.GetTrinoSchema()),
	)

	return nil
}

func getClient() *Client {
	clientOnce.Do(func() {
		clientInstance = &Client{}
	})
	return clientInstance
}

func registerHTTPClient() error {
	httpOnce.Do(func() {
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   defaultConnectTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          defaultMaxOpenConns,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: defaultQueryTimeout,
			},
		}
		httpClientErr = trinodriver.RegisterCustomClient(customHTTPClientName, httpClient)
	})
	return httpClientErr
}

func buildConfig() (*trinodriver.Config, error) {
	scheme := "http"
	if config.IsTrinoTLS() {
		scheme = "https"
	}

	host := config.GetTrinoHost()
	port := config.GetTrinoPort()
	user := config.GetTrinoUser()
	password := config.GetTrinoPassword()

	serverURI := fmt.Sprintf("%s://%s@%s:%s", scheme, user, host, port)
	if password != "" {
		serverURI = fmt.Sprintf("%s://%s:%s@%s:%s", scheme, user, password, host, port)
	}

	queryTimeout := defaultQueryTimeout
	cfg := &trinodriver.Config{
		ServerURI:        serverURI,
		Source:           config.GetTrinoSource(),
		Catalog:          config.GetTrinoCatalog(),
		Schema:           config.GetTrinoSchema(),
		AccessToken:      config.GetTrinoAccessToken(),
		CustomClientName: customHTTPClientName,
		QueryTimeout:     &queryTimeout,
	}

	return cfg, nil
}

func openDB(ctx context.Context) (*sql.DB, error) {
	cfg, err := buildConfig()
	if err != nil {
		return nil, err
	}

	dsn, err := cfg.FormatDSN()
	if err != nil {
		return nil, fmt.Errorf("format trino dsn: %w", err)
	}

	db, err := sql.Open("trino", dsn)
	if err != nil {
		return nil, fmt.Errorf("open trino: %w", err)
	}

	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	var pingResult int
	if err := db.QueryRowContext(pingCtx, "SELECT 1").Scan(&pingResult); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("verify trino connection: %w", err)
	}

	return db, nil
}

// Available reports whether Trino is enabled, configured, and connected.
func Available() bool {
	if !config.IsTrinoConfigured() {
		return false
	}
	c := getClient()
	return c.db != nil
}

// Pool returns the underlying sql.DB when connected, or nil.
func Pool() *sql.DB {
	c := getClient()
	return c.db
}

// Query runs a read query against Trino with context cancellation.
func Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db, err := requireDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("trino query failed",
			logger.Error(err),
			logger.String("catalog", config.GetTrinoCatalog()),
			logger.String("schema", config.GetTrinoSchema()),
		)
		return nil, err
	}

	return rows, nil
}

// QueryRow runs a single-row query against Trino.
func QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	db, err := requireDB()
	if err != nil {
		return nil, err
	}
	return db.QueryRowContext(ctx, query, args...), nil
}

// Ping checks Trino connectivity with a lightweight query.
func Ping(ctx context.Context) error {
	db, err := requireDB()
	if err != nil {
		return err
	}
	var one int
	return db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
}

// Close shuts down the Trino pool.
func Close() {
	c := getClient()
	if c.db != nil {
		_ = c.db.Close()
		c.db = nil
	}
}

func requireDB() (*sql.DB, error) {
	if !config.IsTrinoEnabled() {
		return nil, ErrDisabled
	}
	if !config.IsTrinoConfigured() {
		return nil, ErrNotConfigured
	}
	db := Pool()
	if db == nil {
		return nil, ErrNotConnected
	}
	return db, nil
}
