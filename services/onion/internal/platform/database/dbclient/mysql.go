package dbclient

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

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
