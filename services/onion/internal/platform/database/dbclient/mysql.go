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

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

type MySQLClient struct {
	mainDB    *sql.DB
	loggingDB *sql.DB
	replicaDB *sql.DB
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
	if err := checkDSNSafety("database", mainDSN); err != nil {
		return err
	}

	mainDB, err := openDB(ctx, mainDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to main MySQL database: %w", err)
	}

	// Connect is retried in-process after a crash (cmd/api restarts the
	// server), and every later call reports success from the mainDB != nil
	// early return above. A fatal error below therefore must not leave a
	// partially initialized client behind: the next attempt would skip the
	// logging pool entirely and serve traffic with LoggingPool()/
	// ReplicaPool() silently falling back to the primary.
	initialized := false
	defer func() {
		if initialized {
			return
		}
		client.mainDB = nil
		client.loggingDB = nil
		client.replicaDB = nil
		_ = mainDB.Close()
	}()

	client.mainDB = mainDB

	// The logging pool is writable, so a local run pointed at the production
	// logging database could read and mutate live data. Guarded like the main
	// DSN, and hard-blocking like it too, since a logging connection failure is
	// already fatal here.
	if loggingDSN := resolveLoggingDSN(); loggingDSN != "" {
		if err := checkDSNSafety("logging database", loggingDSN); err != nil {
			return err
		}

		loggingDB, loggingErr := openDB(ctx, loggingDSN)
		if loggingErr != nil {
			return fmt.Errorf("failed to connect to logging MySQL database: %w", loggingErr)
		}

		client.loggingDB = loggingDB
	}

	// Read replica for the AI / NL-to-SQL layer, connected with the read-only
	// tenant-scoped user. Optional and fail-soft: if it's not configured or not
	// reachable, AI reads fall back to the primary via ReplicaPool().
	if replicaConfigured() {
		replicaDSN := buildReplicaDSN()

		// Fail-soft, matching the rest of this path: refusing to open the
		// connection already achieves what the guard exists for, and AI reads
		// fall back to the primary. Hard-blocking would turn an optional,
		// deliberately degradable subsystem into a startup-fatal dependency,
		// which is a bigger behavior change than the guard should make.
		// Logged at Error rather than Warn because a production-shaped replica
		// host is a misconfiguration, not a transiently unreachable replica.
		if safetyErr := checkDSNSafety("read replica", replicaDSN); safetyErr != nil {
			logger.Ctx(ctx).Error("skipping read replica; AI reads will use the primary",
				logger.Error(safetyErr))
		} else if replicaDB, repErr := openDB(ctx, replicaDSN); repErr != nil {
			logger.Ctx(ctx).Warn("failed to connect to read replica; AI reads will use the primary",
				logger.Error(repErr))
		} else {
			client.replicaDB = replicaDB
			logger.Ctx(ctx).Info("Connected to read replica",
				logger.String("host", os.Getenv(envReplicaHost)))
		}
	}

	// Report the host/db actually connected to. Prod sets only DATABASE_URL (no
	// DB_HOST/DB_NAME), so the env-default fallback would otherwise mislabel a
	// real connection as localhost/onion_dev. Parse the resolved DSN first; fall
	// back to env only when it can't be parsed (or for the local-dev DSN).
	logHost := hostFromDSN(mainDSN)
	if logHost == "" {
		logHost = getEnvOrDefault("DB_HOST", DefaultDBHost) + ":" + getEnvOrDefault("DB_PORT", DefaultDBPort)
	}
	logDB := dbNameFromDSN(mainDSN)
	if logDB == "" {
		logDB = getEnvOrDefault("DB_NAME", DefaultDBName)
	}
	logger.Ctx(ctx).Info("Connected to MySQL database",
		logger.String("host", logHost),
		logger.String("database", logDB),
	)

	initialized = true

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
	var remoteDSN string
	if config.GlobalConfig != nil {
		remoteDSN = config.GlobalConfig.DatabaseURL
	} else {
		remoteDSN = os.Getenv("DATABASE_URL")
	}

	if remoteDSN != "" {
		return remoteDSN
	}

	return buildLocalDSN()
}

// resolveLoggingDSN returns the optional logging-database DSN, or "" when none
// is configured. Split out of Connect so the safety guard can be exercised
// against it directly.
func resolveLoggingDSN() string {
	if config.GlobalConfig != nil {
		return config.GlobalConfig.LoggingDatabaseURL
	}

	return os.Getenv("LOGGING_DATABASE_URL")
}

// Read-replica / AI read-only connection env vars.
const (
	envReplicaHost = "DB_REPLICA_HOST"
	envReplicaPort = "DB_REPLICA_PORT"
	envAIUser      = "DB_AI_USER"
	envAIPassword  = "DB_AI_PASSWORD"
)

// replicaConfigured reports whether a read replica + AI read-only credentials are
// available. All three are required; otherwise AI reads fall back to the primary.
func replicaConfigured() bool {
	return os.Getenv(envReplicaHost) != "" && os.Getenv(envAIUser) != "" && os.Getenv(envAIPassword) != ""
}

// replicaDBName resolves the database name for the replica connection. Prod sets
// only DATABASE_URL (no DB_NAME), so when DB_NAME is absent we parse the db name
// out of the primary DSN rather than falling back to the dev default.
func replicaDBName() string {
	if n := os.Getenv("DB_NAME"); n != "" {
		return n
	}
	if n := dbNameFromDSN(resolveMainDSN()); n != "" {
		return n
	}
	return DefaultDBName
}

// hostFromDSN extracts host[:port] from a go-sql-driver DSN
// (user:pass@tcp(host:port)/dbname?params) or a URL DSN
// (scheme://user:pass@host:port/dbname?params). Returns "" if it can't parse,
// so callers can fall back to the DB_HOST env. Used only for log output — it
// never carries credentials.
func hostFromDSN(dsn string) string {
	if i := strings.IndexByte(dsn, '?'); i >= 0 {
		dsn = dsn[:i]
	}
	// native form: ...@tcp(host:port)/...
	if i := strings.Index(dsn, "tcp("); i >= 0 {
		rest := dsn[i+len("tcp("):]
		if j := strings.IndexByte(rest, ')'); j >= 0 {
			return rest[:j]
		}
	}
	// URL form: scheme://[user:pass@]host:port/db
	if i := strings.Index(dsn, "://"); i >= 0 {
		rest := dsn[i+len("://"):]
		if s := strings.IndexByte(rest, '/'); s >= 0 {
			rest = rest[:s]
		}
		if at := strings.LastIndexByte(rest, '@'); at >= 0 {
			rest = rest[at+1:]
		}
		return rest
	}
	return ""
}

// dbNameFromDSN extracts the database name from a go-sql-driver DSN
// (user:pass@tcp(host:port)/dbname?params) or a URL DSN (scheme://.../dbname?params).
func dbNameFromDSN(dsn string) string {
	if i := strings.IndexByte(dsn, '?'); i >= 0 {
		dsn = dsn[:i]
	}
	if i := strings.LastIndexByte(dsn, '/'); i >= 0 {
		return dsn[i+1:]
	}
	return ""
}

// buildReplicaDSN builds the DSN for the AI read-only user against the replica.
// Note: no multiStatements — the NL-to-SQL layer must only ever run a single
// statement, so we disable stacked queries at the connection level too.
func buildReplicaDSN() string {
	host := os.Getenv(envReplicaHost)
	port := getEnvOrDefault(envReplicaPort, getEnvOrDefault("DB_PORT", DefaultDBPort))
	dbname := replicaDBName()
	user := os.Getenv(envAIUser)
	password := os.Getenv(envAIPassword)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&time_zone=%%27UTC%%27&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=%s&readTimeout=%s&writeTimeout=%s&interpolateParams=true",
		user, password, host, port, dbname,
		dsnConnectTimeout, dsnReadTimeout, dsnWriteTimeout,
	)
}

func buildLocalDSN() string {
	host := getEnvOrDefault("DB_HOST", DefaultDBHost)
	port := getEnvOrDefault("DB_PORT", DefaultDBPort)
	dbname := getEnvOrDefault("DB_NAME", DefaultDBName)
	user := getEnvOrDefault("DB_USER", DefaultDBUser)
	password := getEnvOrDefault("DB_PASSWORD", DefaultDBPassword)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&time_zone=%%27UTC%%27&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=%s&readTimeout=%s&writeTimeout=%s&interpolateParams=true",
		user, password, host, port, dbname,
		dsnConnectTimeout, dsnReadTimeout, dsnWriteTimeout,
	)
}

func normalizeDSN(dsn string) string {
	if !strings.Contains(dsn, "://") {
		if strings.Contains(dsn, "?") {
			if !strings.Contains(dsn, "time_zone=") {
				dsn += "&time_zone=%27UTC%27"
			}
			if !strings.Contains(dsn, "parseTime=") {
				dsn += "&parseTime=true"
			}
			if !strings.Contains(dsn, "loc=") {
				dsn += "&loc=UTC"
			}
			return dsn
		}

		return dsn + "?parseTime=true&loc=UTC&time_zone=%27UTC%27"
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
	if query.Get("time_zone") == "" {
		query.Set("time_zone", "'UTC'")
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, password, host, dbname, query.Encode())
}

func intEnvOrDefault(key string, def int) int {
	if config.GlobalConfig != nil {
		switch key {
		case "DB_MAX_OPEN_CONNS":
			if config.GlobalConfig.DBMaxOpenConns > 0 {
				return config.GlobalConfig.DBMaxOpenConns
			}
		case "DB_MAX_IDLE_CONNS":
			if config.GlobalConfig.DBMaxIdleConns > 0 {
				return config.GlobalConfig.DBMaxIdleConns
			}
		}
	}
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
	if config.GlobalConfig != nil {
		switch key {
		case "DB_CONN_MAX_LIFETIME":
			if config.GlobalConfig.DBConnMaxLifetime > 0 {
				return config.GlobalConfig.DBConnMaxLifetime
			}
		case "DB_CONN_MAX_IDLE_TIME":
			if config.GlobalConfig.DBConnMaxIdleTime > 0 {
				return config.GlobalConfig.DBConnMaxIdleTime
			}
		}
	}
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
	if config.GlobalConfig != nil {
		switch key {
		case "DB_HOST":
			if config.GlobalConfig.DBHost != "" {
				return config.GlobalConfig.DBHost
			}
		case "DB_PORT":
			if config.GlobalConfig.DBPort != "" {
				return config.GlobalConfig.DBPort
			}
		case "DB_NAME":
			if config.GlobalConfig.DBName != "" {
				return config.GlobalConfig.DBName
			}
		case "DB_USER":
			if config.GlobalConfig.DBUser != "" {
				return config.GlobalConfig.DBUser
			}
		case "DB_PASSWORD":
			if config.GlobalConfig.DBPass != "" {
				return config.GlobalConfig.DBPass
			}
		}
	}
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

// ReplicaPool returns the read-only replica pool used by the AI / NL-to-SQL
// layer. Falls back to the primary when no replica is configured/reachable.
func ReplicaPool() *sql.DB {
	client := getClient()
	if client.replicaDB != nil {
		return client.replicaDB
	}
	return client.mainDB
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

	if client.replicaDB != nil {
		_ = client.replicaDB.Close()
	}
}

var ErrDatabaseNotReady = errors.New("database not ready after timeout")
