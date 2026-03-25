package dbclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/voxire/onion/internal/platform/logger"
)

type PostgresClient struct {
	voxirePool   *pgxpool.Pool
	contentPool  *pgxpool.Pool
	internalPool *pgxpool.Pool
	loggingPool  *pgxpool.Pool
}

var (
	clientInstance *PostgresClient
	clientOnce     sync.Once
)

const (
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "5432"
	DefaultDBName     = "onion_dev"
	DefaultDBUser     = "onion_user"
	DefaultDBPassword = "onion_password"

	DefaultMaxConnIdleTime = 30 * time.Minute
	DefaultRetryInterval   = 2 * time.Second
)

func getClient() *PostgresClient {
	clientOnce.Do(func() {
		clientInstance = &PostgresClient{}
	})

	return clientInstance
}

func Connect(ctx context.Context) error {
	client := getClient()

	var err error

	connOnce := sync.Once{}
	connOnce.Do(func() {
		if remoteDSN := getEnvOrDefault("DATABASE_URL", ""); remoteDSN != "" {
			logger.Ctx(ctx).Info("Attempting to connect to remote PostgreSQL databases")
			err = client.connectToRemoteDBs(ctx, remoteDSN)
		} else {
			logger.Ctx(ctx).Info("Attempting to connect to local PostgreSQL Docker databases")
			err = client.connectToLocalDBs(ctx)
		}
	})

	return err
}

func (c *PostgresClient) connectToRemoteDBs(ctx context.Context, dsn string) error {
	// Connect to main database
	mainConfig, configErr := pgxpool.ParseConfig(dsn)
	if configErr != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", configErr)
	}

	mainPool, connErr := pgxpool.NewWithConfig(ctx, mainConfig)
	if connErr != nil {
		return fmt.Errorf("failed to connect to main remote database: %w", connErr)
	}

	c.voxirePool = mainPool

	// Connect to content and internal databases
	baseURL := dsn[:strings.LastIndex(dsn, "/")+1]
	queryParams := dsn[strings.LastIndex(dsn, "?"):]

	// Content database
	contentDSN := baseURL + "content" + queryParams

	contentConfig, contentErr := pgxpool.ParseConfig(contentDSN)
	if contentErr != nil {
		return fmt.Errorf("failed to parse content DATABASE_URL: %w", contentErr)
	}

	c.contentPool, contentErr = pgxpool.NewWithConfig(ctx, contentConfig)
	if contentErr != nil {
		return fmt.Errorf("failed to connect to content database: %w", contentErr)
	}

	// Internal database
	internalDSN := baseURL + "internal" + queryParams

	internalConfig, internalErr := pgxpool.ParseConfig(internalDSN)
	if internalErr != nil {
		return fmt.Errorf("failed to parse internal DATABASE_URL: %w", internalErr)
	}

	c.internalPool, internalErr = pgxpool.NewWithConfig(ctx, internalConfig)
	if internalErr != nil {
		return fmt.Errorf("failed to connect to internal database: %w", internalErr)
	}

	// Connect to logging database (optional)
	if loggingDSN := getEnvOrDefault("LOGGING_DATABASE_URL", ""); loggingDSN != "" {
		loggingConfig, loggingErr := pgxpool.ParseConfig(loggingDSN)
		if loggingErr != nil {
			return fmt.Errorf("failed to parse LOGGING_DATABASE_URL: %w", loggingErr)
		}

		c.loggingPool, loggingErr = pgxpool.NewWithConfig(ctx, loggingConfig)
		if loggingErr != nil {
			return fmt.Errorf("failed to connect to logging database: %w", loggingErr)
		}
	}

	logger.Ctx(ctx).Info("Connected to Remote PostgreSQL Databases (voxire + content + internal)")

	return nil
}

func (c *PostgresClient) connectToLocalDBs(ctx context.Context) error {
	// Connect to main database
	mainDSN := buildDockerDSN()

	mainConfig, configErr := pgxpool.ParseConfig(mainDSN)
	if configErr != nil {
		return fmt.Errorf("failed to parse main local DSN: %w", configErr)
	}

	mainConfig.MaxConns = 10
	mainConfig.MinConns = 2
	mainConfig.MaxConnLifetime = 1 * time.Hour
	mainConfig.MaxConnIdleTime = DefaultMaxConnIdleTime
	mainConfig.HealthCheckPeriod = 1 * time.Minute

	mainPool, connErr := pgxpool.NewWithConfig(ctx, mainConfig)
	if connErr != nil {
		return fmt.Errorf("failed to connect to main local Docker database: %w", connErr)
	}

	c.voxirePool = mainPool

	// Connect to logging database (use separate DB name if specified)
	loggingDBName := getEnvOrDefault("LOGGING_DB_NAME", getEnvOrDefault("DB_NAME", DefaultDBName))
	loggingDSN := buildDockerDSNWithDBName(loggingDBName)

	loggingConfig, loggingConfigErr := pgxpool.ParseConfig(loggingDSN)
	if loggingConfigErr != nil {
		return fmt.Errorf("failed to parse logging local DSN: %w", loggingConfigErr)
	}

	loggingConfig.MaxConns = 5
	loggingConfig.MinConns = 1
	loggingConfig.MaxConnLifetime = 1 * time.Hour
	loggingConfig.MaxConnIdleTime = DefaultMaxConnIdleTime
	loggingConfig.HealthCheckPeriod = 1 * time.Minute

	loggingPool, loggingConnErr := pgxpool.NewWithConfig(ctx, loggingConfig)
	if loggingConnErr != nil {
		return fmt.Errorf("failed to connect to logging local Docker database: %w", loggingConnErr)
	}

	c.loggingPool = loggingPool

	host := getEnvOrDefault("DB_HOST", DefaultDBHost)
	port := getEnvOrDefault("DB_PORT", DefaultDBPort)
	mainDBName := getEnvOrDefault("DB_NAME", DefaultDBName)
	logger.Ctx(ctx).Info("Connected to Local PostgreSQL Docker Databases",
		logger.String("host", host),
		logger.String("port", port),
		logger.String("main_database", mainDBName),
		logger.String("logging_database", loggingDBName))

	return nil
}

func buildDockerDSN() string {
	host := getEnvOrDefault("DB_HOST", DefaultDBHost)
	port := getEnvOrDefault("DB_PORT", DefaultDBPort)
	dbname := getEnvOrDefault("DB_NAME", DefaultDBName)
	user := getEnvOrDefault("DB_USER", DefaultDBUser)
	password := getEnvOrDefault("DB_PASSWORD", DefaultDBPassword)

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}

func buildDockerDSNWithDBName(dbname string) string {
	host := getEnvOrDefault("DB_HOST", DefaultDBHost)
	port := getEnvOrDefault("DB_PORT", DefaultDBPort)
	user := getEnvOrDefault("DB_USER", DefaultDBUser)
	password := getEnvOrDefault("DB_PASSWORD", DefaultDBPassword)

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
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
	if !arePoolsInitialized(client) {
		return false
	}

	return arePoolsHealthy(ctx, client)
}

func arePoolsInitialized(client *PostgresClient) bool {
	return client.voxirePool != nil && client.contentPool != nil && client.internalPool != nil
}

func arePoolsHealthy(ctx context.Context, client *PostgresClient) bool {
	if err := client.voxirePool.Ping(ctx); err != nil {
		return false
	}

	if err := client.contentPool.Ping(ctx); err != nil {
		return false
	}

	if err := client.internalPool.Ping(ctx); err != nil {
		return false
	}

	return true
}

type DBManager struct {
	MainPool    *pgxpool.Pool
	LoggingPool *pgxpool.Pool
}

func VoxirePool() *pgxpool.Pool {
	client := getClient()

	return client.voxirePool
}

func ContentPool() *pgxpool.Pool {
	client := getClient()

	return client.contentPool
}

func InternalPool() *pgxpool.Pool {
	client := getClient()

	return client.internalPool
}

func LoggingPool() *pgxpool.Pool {
	client := getClient()

	return client.loggingPool
}

// Pool returns VoxirePool for backward compatibility.
func Pool() *pgxpool.Pool {
	return VoxirePool()
}

func GetPoolByName(dbName string) *pgxpool.Pool {
	switch dbName {
	case "voxire":
		return VoxirePool()
	case "content":
		return ContentPool()
	case "internal":
		return InternalPool()
	default:
		return VoxirePool()
	}
}

func Close() {
	client := getClient()

	if client.voxirePool != nil {
		client.voxirePool.Close()
	}

	if client.contentPool != nil {
		client.contentPool.Close()
	}

	if client.internalPool != nil {
		client.internalPool.Close()
	}

	if client.loggingPool != nil {
		client.loggingPool.Close()
	}
}

var ErrDatabaseNotReady = errors.New("database not ready after timeout")
