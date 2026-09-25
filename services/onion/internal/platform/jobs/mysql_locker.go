package jobs

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

const defaultMySQLLockPrefix = "onion.job."

type MySQLLocker struct {
	db     *sql.DB
	prefix string
}

func NewMySQLLocker(db *sql.DB) *MySQLLocker {
	return &MySQLLocker{db: db, prefix: defaultMySQLLockPrefix}
}

func (l *MySQLLocker) TryLock(ctx context.Context, name string) (func(context.Context) error, bool, error) {
	if l == nil || l.db == nil {
		return nil, true, nil
	}

	conn, err := l.db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}

	lockName := l.lockName(name)
	var acquired sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 0)", lockName).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, false, err
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		_ = conn.Close()
		return nil, false, nil
	}

	var once sync.Once
	release := func(ctx context.Context) error {
		var err error
		once.Do(func() {
			var released sql.NullInt64
			err = conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", lockName).Scan(&released)
			if err == nil && released.Valid && released.Int64 == 0 {
				err = fmt.Errorf("mysql lock %q was not held by this connection", lockName)
			}
			if closeErr := conn.Close(); err == nil {
				err = closeErr
			}
		})
		return err
	}
	return release, true, nil
}

func (l *MySQLLocker) lockName(name string) string {
	prefix := strings.TrimSpace(l.prefix)
	if prefix == "" {
		prefix = defaultMySQLLockPrefix
	}
	full := prefix + name
	if len(full) <= 64 {
		return full
	}
	sum := sha256.Sum256([]byte(full))
	hexSum := hex.EncodeToString(sum[:])
	space := 64 - len(prefix)
	if space <= 0 {
		return hexSum[:64]
	}
	return prefix + hexSum[:space]
}
