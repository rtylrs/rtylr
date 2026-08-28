package containers

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // register MySQL database/sql driver for test containers
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MySQLConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	DSN      string
}

func StartMySQL(ctx context.Context, database, migrationsDir, seedsDir string) (MySQLConfig, func(context.Context) error, error) {
	if database == "" {
		database = "onion_test"
	}

	const user = "testuser"
	const pass = "testpass"
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": pass,
			"MYSQL_DATABASE":      database,
			"MYSQL_USER":          user,
			"MYSQL_PASSWORD":      pass,
		},
		WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return MySQLConfig{}, nil, err
	}

	cleanup := func(ctx context.Context) error {
		return container.Terminate(ctx)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = cleanup(ctx)
		return MySQLConfig{}, nil, err
	}
	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		_ = cleanup(ctx)
		return MySQLConfig{}, nil, err
	}

	cfg := MySQLConfig{
		Host:     host,
		Port:     port.Port(),
		Database: database,
		User:     user,
		Password: pass,
	}
	cfg.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&multiStatements=true&charset=utf8mb4&timeout=10s&readTimeout=15s&writeTimeout=15s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	if err := applySQLDir(ctx, cfg.DSN, migrationsDir); err != nil {
		_ = cleanup(ctx)
		return MySQLConfig{}, nil, err
	}
	if err := applySQLDir(ctx, cfg.DSN, seedsDir); err != nil {
		_ = cleanup(ctx)
		return MySQLConfig{}, nil, err
	}

	return cfg, cleanup, nil
}

func applySQLDir(ctx context.Context, dsn, dir string) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	files, err := sqlFiles(dir)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, file := range files {
		// #nosec G304 -- Test setup reads the validated migration files returned by sqlFiles.
		body, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}
	}
	return nil
}

func sqlFiles(dir string) ([]string, error) {
	if !filepath.IsAbs(dir) {
		root, err := projectRoot()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(root, dir)
	}

	var files []string
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = next
	}
}
