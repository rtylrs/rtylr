package dbclient

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

	DefaultConnMaxIdleTime = 30 * time.Minute
	DefaultRetryInterval   = 2 * time.Second
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

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(DefaultConnMaxIdleTime)
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()

		return nil, err
	}

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

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user,
		password,
		host,
		port,
		dbname,
	)
}

func normalizeDSN(dsn string) string {
	if !strings.Contains(dsn, "://") {
		if strings.Contains(dsn, "parseTime=") {
			return dsn
		}

		if strings.Contains(dsn, "?") {
			return dsn + "&parseTime=true"
		}

		return dsn + "?parseTime=true"
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

	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, password, host, dbname, query.Encode())
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
