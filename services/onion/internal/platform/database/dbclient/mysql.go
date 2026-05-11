package dbclient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // register MySQL driver
	"github.com/voxire/onion/internal/platform/logger"
)

type MySQLClient struct {
	mainDB    *sql.DB
	loggingDB *sql.DB
}

var (
	clientInstance *MySQLClient
	clientOnce     sync.Once
)

const (
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "3306"
	DefaultDBName     = "onion_dev"
	DefaultDBUser     = "onion_user"
	DefaultDBPassword = "onion_password"

	// Pool defaults — tuned for a production SaaS with ~1 000 users.
	// MySQL's own max_connections is typically 100-300 on shared/small
	// hosts. We stay well below that so other processes (migrations,
	// monitoring, ad-hoc queries) always have headroom.
	//
	// Override via env vars:
	//   DB_MAX_OPEN_CONNS   (default 25)
	//   DB_MAX_IDLE_CONNS   (default 10)
	//   DB_CONN_MAX_LIFETIME (default 30m, e.g. "30m", "1h")
	//   DB_CONN_MAX_IDLE_TIME (default 10m)
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 10
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultConnMaxIdleTime = 10 * time.Minute
	DefaultRetryInterval   = 2 * time.Second

	// DSN-level timeouts — prevent a single slow query or dead socket
	// from hanging a goroutine indefinitely.
	// readTimeout / writeTimeout are the network I/O deadlines; any
	// query that hasn't finished sending/receiving within this window
	// gets a hard socket close. Keep in sync with the HTTP server's
	// WriteTimeout and the request timeout middleware.
	dsnConnectTimeout = "10s"
	dsnReadTimeout    = "15s"
	dsnWriteTimeout   = "15s"
)

func getClient() *MySQLClient {
	clientOnce.Do(func() {
		clientInstance = &MySQLClient{}
	})

	return clientInstance
}

func Connect(ctx context.Context) error {
	client := getClient()

	if client.mainDB != nil {
		return nil
	}

	mainDSN := resolveMainDSN()
	mainDB, err := openDB(ctx, mainDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to main MySQL database: %w", err)
	}

	client.mainDB = mainDB

	loggingDSN := os.Getenv("LOGGING_DATABASE_URL")
	if loggingDSN != "" {
		loggingDB, loggingErr := openDB(ctx, loggingDSN)
		if loggingErr != nil {
			return fmt.Errorf("failed to connect to logging MySQL database: %w", loggingErr)
		}

		client.loggingDB = loggingDB
	}

	logger.Ctx(ctx).Info("Connected to MySQL database",
		logger.String("host", getEnvOrDefault("DB_HOST", DefaultDBHost)),
		logger.String("port", getEnvOrDefault("DB_PORT", DefaultDBPort)),
		logger.String("database", getEnvOrDefault("DB_NAME", DefaultDBName)),
	)

	return nil
}

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", normalizeDSN(dsn))
	if err != nil {
		return nil, err
	}

	maxOpen := intEnvOrDefault("DB_MAX_OPEN_CONNS", DefaultMaxOpenConns)
	maxIdle := intEnvOrDefault("DB_MAX_IDLE_CONNS", DefaultMaxIdleConns)
	lifetime := durationEnvOrDefault("DB_CONN_MAX_LIFETIME", DefaultConnMaxLifetime)
	idleTime := durationEnvOrDefault("DB_CONN_MAX_IDLE_TIME", DefaultConnMaxIdleTime)

	// Clamp idle to open — having more idle than open makes no sense
	// and would cause sql/database to waste resources.
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)
	db.SetConnMaxIdleTime(idleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	stats := db.Stats()
	logger.Ctx(ctx).Info("database pool configured",
		logger.Int("max_open", maxOpen),
		logger.Int("max_idle", maxIdle),
		logger.String("conn_max_lifetime", lifetime.String()),
		logger.String("conn_max_idle_time", idleTime.String()),
		logger.Int("open_connections", stats.OpenConnections),
	)

	return db, nil
}

func resolveMainDSN() string {
	if remoteDSN := os.Getenv("DATABASE_URL"); remoteDSN != "" {
		return remoteDSN
	}

	return buildLocalDSN()
}

func buildLocalDSN() string {
	host := getEnvOrDefault("DB_HOST", DefaultDBHost)
	port := getEnvOrDefault("DB_PORT", DefaultDBPort)
	dbname := getEnvOrDefault("DB_NAME", DefaultDBName)
	user := getEnvOrDefault("DB_USER", DefaultDBUser)
	password := getEnvOrDefault("DB_PASSWORD", DefaultDBPassword)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=%s&readTimeout=%s&writeTimeout=%s&interpolateParams=true",
		user, password, host, port, dbname,
		dsnConnectTimeout, dsnReadTimeout, dsnWriteTimeout,
	)
}

func normalizeDSN(dsn string) string {
	if !strings.Contains(dsn, "://") {
		if strings.Contains(dsn, "parseTime=") {
			return dsn
		}

		if strings.Contains(dsn, "?") {
			return dsn + "&parseTime=true&loc=UTC"
		}

		return dsn + "?parseTime=true&loc=UTC"
	}

	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "mysql" {
		return dsn
	}

	user := parsed.User.Username()
	password, _ := parsed.User.Password()
	host := parsed.Host
	dbname := strings.TrimPrefix(parsed.Path, "/")

	query := parsed.Query()
	if query.Get("parseTime") == "" {
		query.Set("parseTime", "true")
	}
	if query.Get("charset") == "" {
		query.Set("charset", "utf8mb4")
	}
	if query.Get("collation") == "" {
		query.Set("collation", "utf8mb4_unicode_ci")
	}
	if query.Get("timeout") == "" {
		query.Set("timeout", dsnConnectTimeout)
	}
	if query.Get("readTimeout") == "" {
		query.Set("readTimeout", dsnReadTimeout)
	}
	if query.Get("writeTimeout") == "" {
		query.Set("writeTimeout", dsnWriteTimeout)
	}
	if query.Get("interpolateParams") == "" {
		query.Set("interpolateParams", "true")
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, password, host, dbname, query.Encode())
}

func intEnvOrDefault(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func durationEnvOrDefault(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func WaitForDB(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if isDBReady(ctx) {
			return nil
		}

		time.Sleep(DefaultRetryInterval)
	}

	return ErrDatabaseNotReady
}

func isDBReady(ctx context.Context) bool {
	if err := Connect(ctx); err != nil {
		return false
	}

	client := getClient()
	if client.mainDB == nil {
		return false
	}

	return client.mainDB.PingContext(ctx) == nil
}

type DBManager struct {
	MainDB    *sql.DB
	LoggingDB *sql.DB
}

func VoxirePool() *sql.DB {
	return Pool()
}

func ContentPool() *sql.DB {
	return Pool()
}

func InternalPool() *sql.DB {
	return Pool()
}

func LoggingPool() *sql.DB {
	client := getClient()
	if client.loggingDB != nil {
		return client.loggingDB
	}

	return client.mainDB
}

func Pool() *sql.DB {
	return getClient().mainDB
}

func GetPoolByName(_ string) *sql.DB {
	return Pool()
}

func Close() {
	client := getClient()

	if client.mainDB != nil {
		_ = client.mainDB.Close()
	}

	if client.loggingDB != nil {
		_ = client.loggingDB.Close()
	}
}

var ErrDatabaseNotReady = errors.New("database not ready after timeout")
