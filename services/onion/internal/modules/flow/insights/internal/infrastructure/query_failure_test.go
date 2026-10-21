package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/testutil"
)

var errInjectedQueryFailure = errors.New("injected query failure")

// failConnectDriver fails every connection attempt. Opening a *sql.DB on it
// yields a *sql.Row whose Scan reports the failure, which is how the stub below
// makes one generated single-row query fail without a query-mocking dependency.
type failConnectDriver struct{}

func (failConnectDriver) Open(string) (driver.Conn, error) { return nil, errInjectedQueryFailure }

func (failConnectDriver) Connect(context.Context) (driver.Conn, error) {
	return nil, errInjectedQueryFailure
}

func (failConnectDriver) Driver() driver.Driver { return failConnectDriver{} }

// failMatchingQueriesDBTX delegates to a real handle, except for the generated
// query whose name contains fragment, which reports errInjectedQueryFailure.
// That models a transient failure of one secondary insights query while the
// rest of the request would otherwise succeed.
type failMatchingQueriesDBTX struct {
	real         sqlc.DBTX
	fragment     string
	failingRowDB *sql.DB
}

func (f *failMatchingQueriesDBTX) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return f.real.ExecContext(ctx, query, args...)
}

func (f *failMatchingQueriesDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return f.real.PrepareContext(ctx, query)
}

func (f *failMatchingQueriesDBTX) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if strings.Contains(query, f.fragment) {
		return nil, errInjectedQueryFailure
	}
	return f.real.QueryContext(ctx, query, args...)
}

func (f *failMatchingQueriesDBTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if strings.Contains(query, f.fragment) {
		return f.failingRowDB.QueryRowContext(ctx, query, args...)
	}
	return f.real.QueryRowContext(ctx, query, args...)
}

func newFailingQueryRepository(t *testing.T, fragment string) *Repository {
	t.Helper()

	db := testutil.OpenDB(t)
	failingRowDB := sql.OpenDB(failConnectDriver{})
	t.Cleanup(func() { _ = failingRowDB.Close() })

	repo := NewRepository(db)
	repo.queries = sqlc.New(&failMatchingQueriesDBTX{
		real:         db,
		fragment:     fragment,
		failingRowDB: failingRowDB,
	})
	return repo
}

// A transient failure of a secondary single-row stats query must surface as an
// error. Returning the zeroed field as success made insights.dashboard.v1
// answer 200 with zero velocity and the application layer cache the false
// zeros for two minutes.
func TestGetInsightsStats_VelocityQueryFailureIsReturned(t *testing.T) {
	repo := newFailingQueryRepository(t, "GetVelocityThisWeek")

	if _, err := repo.GetInsightsStats(context.Background(), testutil.OwnerID); !errors.Is(err, errInjectedQueryFailure) {
		t.Fatalf("GetInsightsStats error = %v, want the injected secondary-query failure", err)
	}
}

func TestGetInsightsStats_RowQueryFailureIsReturned(t *testing.T) {
	repo := newFailingQueryRepository(t, "GetTasksByStatus")

	if _, err := repo.GetInsightsStats(context.Background(), testutil.OwnerID); !errors.Is(err, errInjectedQueryFailure) {
		t.Fatalf("GetInsightsStats error = %v, want the injected secondary-query failure", err)
	}
}

// The milestones query is a separate statement from the projects query, so a
// failure there used to come back as a successful roadmap whose milestone
// lists were all empty. Distinct ids keep the seeded project away from the
// shared fixture and other tests.
func TestGetRoadmap_MilestonesQueryFailureIsReturned(t *testing.T) {
	const (
		ownerID    int64 = 941500
		businessID int64 = 941600
	)

	db := testutil.OpenDB(t)
	seedFailingQueryRoadmapProject(t, db, ownerID, businessID)

	failingRowDB := sql.OpenDB(failConnectDriver{})
	t.Cleanup(func() { _ = failingRowDB.Close() })

	repo := NewRepository(db)
	repo.queries = sqlc.New(&failMatchingQueriesDBTX{
		real:         db,
		fragment:     "GetMilestonesByProjectIDs",
		failingRowDB: failingRowDB,
	})

	if _, err := repo.GetRoadmap(context.Background(), businessID); !errors.Is(err, errInjectedQueryFailure) {
		t.Fatalf("GetRoadmap error = %v, want the injected milestones failure", err)
	}
}

func seedFailingQueryRoadmapProject(t *testing.T, db *sql.DB, ownerID, businessID int64) {
	t.Helper()

	ctx := context.Background()
	res, err := db.ExecContext(ctx,
		`INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, status, start_date, due_date, created_by)
		 VALUES (0, ?, ?, 'Failure-injection project', 'active', '2027-01-01', '2027-12-31', ?)`,
		ownerID, businessID, ownerID,
	)
	if err != nil {
		t.Fatalf("insert flow_projects: %v", err)
	}
	projectID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projectID)
	})
}
