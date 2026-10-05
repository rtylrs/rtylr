package infrastructure

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

// Only a unique-key violation is worth retrying. Anything else — a deadlock, a
// connection reset, a validation failure — must surface immediately, because
// re-running the transaction would only fail the same way.
func TestIsDuplicateEntry(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "a unique key violation is retryable",
			err:  &mysql.MySQLError{Number: mysqlErrDuplicateEntry, Message: "Duplicate entry"},
			want: true,
		},
		{
			name: "a wrapped unique key violation is retryable",
			err:  fmt.Errorf("insert pos_orders: %w", &mysql.MySQLError{Number: mysqlErrDuplicateEntry}),
			want: true,
		},
		{
			name: "a deadlock is not retryable",
			err:  &mysql.MySQLError{Number: 1213, Message: "Deadlock found"},
			want: false,
		},
		{
			name: "a lock wait timeout is not retryable",
			err:  &mysql.MySQLError{Number: 1205},
			want: false,
		},
		{
			name: "a plain error is not retryable",
			err:  errors.New("connection reset"),
			want: false,
		},
		{
			name: "nil is not retryable",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDuplicateEntry(tc.err); got != tc.want {
				t.Fatalf("isDuplicateEntry(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// The retry is bounded: a create that keeps colliding must give up rather than
// spin, and it must give up with the duplicate error so the caller can tell
// what happened.
func TestCreateOrderNumberAttemptsIsBounded(t *testing.T) {
	if createOrderNumberAttempts < 2 {
		t.Fatalf("createOrderNumberAttempts = %d, want at least 2 so a lost race is retried once",
			createOrderNumberAttempts)
	}
	if createOrderNumberAttempts > 10 {
		t.Fatalf("createOrderNumberAttempts = %d, want a small bound; a permanently colliding key would hold a transaction open",
			createOrderNumberAttempts)
	}
}
