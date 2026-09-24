package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
)

// emptyRowsDriver answers every query with zero rows, which is what the
// labour-cost aggregate returns for a project that has no billable time.
type emptyRowsDriver struct{}

func (emptyRowsDriver) Open(string) (driver.Conn, error) { return emptyRowsConn{}, nil }

type emptyRowsConn struct{}

func (emptyRowsConn) Prepare(string) (driver.Stmt, error) { return emptyRowsStmt{}, nil }
func (emptyRowsConn) Close() error                        { return nil }
func (emptyRowsConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

type emptyRowsStmt struct{}

func (emptyRowsStmt) Close() error                               { return nil }
func (emptyRowsStmt) NumInput() int                              { return -1 }
func (emptyRowsStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (emptyRowsStmt) Query([]driver.Value) (driver.Rows, error)  { return emptyRows{}, nil }

type emptyRows struct{}

func (emptyRows) Columns() []string         { return []string{"business_id"} }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func init() { sql.Register("hr-payroll-empty-rows", emptyRowsDriver{}) }

func TestGetProjectLabourCostTreatsNoRowsAsZeroCost(t *testing.T) {
	db, err := sql.Open("hr-payroll-empty-rows", "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cost, err := NewRepository(db).GetProjectLabourCost(context.Background(), 6, 28)
	if err != nil {
		t.Fatalf("GetProjectLabourCost with no billable rows: %v", err)
	}
	if cost.BusinessID != 6 || cost.ProjectID != 28 {
		t.Fatalf("scope = business %d, project %d; want 6, 28", cost.BusinessID, cost.ProjectID)
	}
	if cost.CostCents != 0 || cost.CostedMinutes != 0 || cost.UncostedMinutes != 0 || cost.UncostedEmployeeCount != 0 {
		t.Fatalf("aggregate = %+v; want all zero", cost)
	}
}
