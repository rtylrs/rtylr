package dbclient

import (
	"database/sql"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql" // parse native DSNs and register the MySQL driver

	"github.com/voxire/onion/internal/platform/config"
)

func normalizeURLDSN(dsn string) string {
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "mysql" || parsed.Host == "" {
		return dsn
	}

	// Parse query options through the driver, while assigning credentials and
	// the address from net/url. This avoids mistaking a password's '@', '/', or
	// '?' for DSN delimiters and keeps the returned native DSN compatible with
	// the driver.
	raw := "placeholder@tcp(" + parsed.Host + ")/placeholder"
	query := parsed.Query()
	if encoded := query.Encode(); encoded != "" {
		raw += "?" + encoded
	}
	cfg, err := mysql.ParseDSN(raw)
	if err != nil {
		return dsn
	}
	cfg.User = ""
	cfg.Passwd = ""
	if parsed.User != nil {
		cfg.User = parsed.User.Username()
		cfg.Passwd, _ = parsed.User.Password()
	}
	escapedPath := strings.TrimPrefix(parsed.EscapedPath(), "/")
	cfg.DBName, err = url.PathUnescape(escapedPath)
	if err != nil {
		return dsn
	}

	if _, present := query["parseTime"]; !present {
		cfg.ParseTime = true
	}
	setDSNParam(cfg, query, "charset", "utf8mb4")
	setDSNParam(cfg, query, "collation", "utf8mb4_unicode_ci")
	setDSNParam(cfg, query, "timeout", dsnConnectTimeout)
	setDSNParam(cfg, query, "readTimeout", dsnReadTimeout)
	setDSNParam(cfg, query, "writeTimeout", dsnWriteTimeout)
	setDSNParam(cfg, query, "interpolateParams", "true")
	setDSNParam(cfg, query, "time_zone", "'UTC'")
	return cfg.FormatDSN()
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
