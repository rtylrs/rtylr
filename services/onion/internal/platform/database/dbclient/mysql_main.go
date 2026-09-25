package dbclient

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql" // parse native DSNs and register the MySQL driver

	"github.com/voxire/onion/internal/platform/config"
)

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

type parsedDSNMetadata struct {
	host   string
	dbName string
}

func isMySQLURLDSN(dsn string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(dsn)), "mysql://")
}

// parseDSNMetadata parses the DSN with the same boundary rules as the MySQL
// driver. In particular, native DSN credentials may legally contain '?', '@',
// '/', and even text such as "tcp(...)"; delimiter searches over the whole DSN
// cannot distinguish those bytes from the real endpoint.
func parseDSNMetadata(dsn string) (parsedDSNMetadata, error) {
	if isMySQLURLDSN(dsn) {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return parsedDSNMetadata{}, err
		}
		if parsed.Scheme != "mysql" || parsed.Host == "" {
			return parsedDSNMetadata{}, fmt.Errorf("unsupported MySQL DSN URL")
		}

		escapedPath := strings.TrimPrefix(parsed.EscapedPath(), "/")
		dbName, err := url.PathUnescape(escapedPath)
		if err != nil {
			return parsedDSNMetadata{}, err
		}
		return parsedDSNMetadata{host: parsed.Host, dbName: dbName}, nil
	}

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return parsedDSNMetadata{}, err
	}
	return parsedDSNMetadata{host: cfg.Addr, dbName: cfg.DBName}, nil
}

// hostFromDSN extracts host[:port] from a go-sql-driver DSN
// (user:pass@tcp(host:port)/dbname?params) or a URL DSN
// (scheme://user:pass@host:port/dbname?params). Returns "" only when the DSN
// cannot be parsed. Used only for log output and safety checks — it never
// carries credentials.
func hostFromDSN(dsn string) string {
	metadata, err := parseDSNMetadata(dsn)
	if err != nil {
		return ""
	}
	return metadata.host
}

// dbNameFromDSN extracts the database name from a go-sql-driver DSN
// (user:pass@tcp(host:port)/dbname?params) or a URL DSN (scheme://.../dbname?params).
func dbNameFromDSN(dsn string) string {
	metadata, err := parseDSNMetadata(dsn)
	if err != nil {
		return ""
	}
	return metadata.dbName
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
	if isMySQLURLDSN(dsn) {
		return normalizeURLDSN(dsn)
	}

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return dsn
	}
	params := nativeDSNParams(dsn)
	if _, present := params["parseTime"]; !present {
		cfg.ParseTime = true
	}
	if _, present := params["loc"]; !present {
		cfg.Loc = time.UTC
	}
	setDSNParam(cfg, params, "time_zone", "'UTC'")
	return cfg.FormatDSN()
}

func nativeDSNParams(dsn string) url.Values {
	slash := strings.LastIndexByte(dsn, '/')
	if slash < 0 || slash+1 >= len(dsn) {
		return url.Values{}
	}
	query := dsn[slash+1:]
	question := strings.IndexByte(query, '?')
	if question < 0 {
		return url.Values{}
	}
	values, err := url.ParseQuery(query[question+1:])
	if err != nil {
		return url.Values{}
	}
	return values
}

func setDSNParam(cfg *mysql.Config, params url.Values, name, value string) {
	if _, present := params[name]; present {
		return
	}
	if cfg.Params == nil {
		cfg.Params = make(map[string]string)
	}
	cfg.Params[name] = value
}
