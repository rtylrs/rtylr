package jobs

import (
	"strings"
	"testing"
)

func TestMySQLLockerLockNameFitsMySQLLimit(t *testing.T) {
	locker := NewMySQLLocker(nil)
	name := locker.lockName(strings.Repeat("very-long-job-name-", 10))
	if len(name) > 64 {
		t.Fatalf("lock name length = %d, want <= 64", len(name))
	}
	if !strings.HasPrefix(name, defaultMySQLLockPrefix) {
		t.Fatalf("lock name = %q, want prefix %q", name, defaultMySQLLockPrefix)
	}
}
