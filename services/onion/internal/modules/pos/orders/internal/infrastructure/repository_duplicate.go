package infrastructure

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysqlErrDuplicateEntry is MySQL's ER_DUP_ENTRY.
const mysqlErrDuplicateEntry = 1062

// isDuplicateEntry reports whether err is a unique-key violation. It is used
// to tell "another transaction took this order number first" (retryable, the
// caller recomputes and tries again) apart from a genuine failure.
func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry
}
