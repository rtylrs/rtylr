package infrastructure

import (
	"errors"
	"strings"
	"testing"

	gosqlmysql "github.com/go-sql-driver/mysql"
)

func TestIsOptionalAnalyticsSchemaError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "unknown column", err: &gosqlmysql.MySQLError{Number: 1054}, want: true},
		{name: "missing table", err: &gosqlmysql.MySQLError{Number: 1146}, want: true},
		{name: "duplicate entry", err: &gosqlmysql.MySQLError{Number: mysqlErrDuplicateEntry}, want: false},
		{name: "generic error", err: errors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOptionalAnalyticsSchemaError(tt.err); got != tt.want {
				t.Fatalf("isOptionalAnalyticsSchemaError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScopedRecruitQueriesRequireCareerOwnerJoin(t *testing.T) {
	queries := []string{
		updateApplicationStatusScopedSQL,
		updateApplicationReviewScopedSQL,
		getApplicationCVKeyScopedSQL,
		getApplicationTimelineMetaScopedSQL,
		listApplicationEventsScopedSQL,
		getTaskByIDScopedSQL,
		updateFollowUpTaskScopedSQL,
		applicationExistsScopedSQL,
	}

	for _, query := range queries {
		lower := strings.ToLower(query)
		if !strings.Contains(lower, "join careers c") || !strings.Contains(lower, "c.owner_id = ?") {
			t.Errorf("scoped query is missing the career owner predicate:\n%s", query)
		}
	}
}
