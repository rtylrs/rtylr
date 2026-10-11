package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

type orderFixture struct {
	db       *sql.DB
	repo     *infrastructure.Repository
	ctx      context.Context
	colA     int64
	colB     int64
	tasksA   []int64 // three tasks in project A, all in column A
	taskB    int64   // a task in another project
	taskBCol int64
}

func newOrderFixture(t *testing.T) *orderFixture {
	t.Helper()
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "ordered-move-ws", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_workspaces WHERE id = ?", wsID) })

	mkProject := func(name string) int64 {
		id, err := repo.CreateProject(ctx, domain.ProjectInput{
			WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Name: name, CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateProject %s: %v", name, err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", id) })
		return id
	}
	mkColumn := func(project int64, name string) int64 {
		id, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
			ProjectID: project, Name: name, CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateColumn %s: %v", name, err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_project_columns WHERE id = ?", id) })
		return id
	}
	mkTask := func(project, column int64, title string) int64 {
		id, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: project, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title: title, ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %s: %v", title, err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE flow_tasks SET section_id = ? WHERE id = ?", column, id); err != nil {
			t.Fatalf("seed section: %v", err)
		}
		t.Cleanup(func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) })
		return id
	}

	projectA := mkProject("ordered-move-a")
	projectB := mkProject("ordered-move-b")
	f := &orderFixture{db: db, repo: repo, ctx: ctx}
	f.colA = mkColumn(projectA, "A1")
	f.colB = mkColumn(projectA, "A2")
	f.taskBCol = mkColumn(projectB, "B1")
	for _, title := range []string{"t1", "t2", "t3"} {
		f.tasksA = append(f.tasksA, mkTask(projectA, f.colA, title))
	}
	f.taskB = mkTask(projectB, f.taskBCol, "other project")
	return f
}

func (f *orderFixture) sortOrder(t *testing.T, id int64) int {
	t.Helper()
	var n int
	if err := f.db.QueryRowContext(f.ctx, "SELECT COALESCE(sort_order,0) FROM flow_tasks WHERE id = ?", id).Scan(&n); err != nil {
		t.Fatalf("read sort_order: %v", err)
	}
	return n
}

func (f *orderFixture) section(t *testing.T, id int64) int64 {
	t.Helper()
	var s sql.NullInt64
	if err := f.db.QueryRowContext(f.ctx, "SELECT section_id FROM flow_tasks WHERE id = ?", id).Scan(&s); err != nil {
		t.Fatalf("read section: %v", err)
	}
	return s.Int64
}

func (f *orderFixture) wantOrders(t *testing.T, ids []int64, want []int) {
	t.Helper()
	for i, id := range ids {
		if got := f.sortOrder(t, id); got != want[i] {
			t.Errorf("task %d sort_order = %d, want %d", id, got, want[i])
		}
	}
}

func TestMoveTaskToColumnOrderedSavesPositions(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	// Reorder within the same column: t3, t1, t2.
	order := []int64{a[2], a[0], a[1]}
	if err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[2], f.colA, testutil.OwnerID, order); err != nil {
		t.Fatalf("MoveTaskToColumnOrdered: %v", err)
	}
	f.wantOrders(t, order, []int{1, 2, 3})
	if got := f.section(t, a[2]); got != f.colA {
		t.Errorf("section = %d, want %d", got, f.colA)
	}
}

func TestMoveTaskToColumnOrderedInsertsIntoAnotherColumn(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	// t1 and t2 already in column B at positions 1,2; drop t3 between them.
	for i, id := range a[:2] {
		if _, err := f.db.ExecContext(f.ctx, "UPDATE flow_tasks SET section_id = ?, sort_order = ? WHERE id = ?", f.colB, i+1, id); err != nil {
			t.Fatal(err)
		}
	}
	order := []int64{a[0], a[2], a[1]}
	if err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[2], f.colB, testutil.OwnerID, order); err != nil {
		t.Fatalf("MoveTaskToColumnOrdered: %v", err)
	}
	if got := f.section(t, a[2]); got != f.colB {
		t.Errorf("section = %d, want %d", got, f.colB)
	}
	f.wantOrders(t, order, []int{1, 2, 3})
}

func TestMoveTaskToColumnOrderedEmptyListKeepsSortOrder(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	if _, err := f.db.ExecContext(f.ctx, "UPDATE flow_tasks SET sort_order = 7 WHERE id IN (?, ?)", a[0], a[1]); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[0], f.colB, testutil.OwnerID, nil); err != nil {
		t.Fatalf("MoveTaskToColumnOrdered: %v", err)
	}
	if got := f.section(t, a[0]); got != f.colB {
		t.Errorf("section = %d, want %d", got, f.colB)
	}
	f.wantOrders(t, a[:2], []int{7, 7})
}

func TestMoveTaskToColumnOrderedRejectsForeignProjectWithoutPartialWrite(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	order := []int64{a[0], f.taskB, a[1]}
	err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[0], f.colB, testutil.OwnerID, order)
	if !errors.Is(err, domain.ErrTaskOrderOutsideProject) {
		t.Fatalf("err = %v, want ErrTaskOrderOutsideProject", err)
	}
	// Rolled back: the move did not stick and no order was written.
	if got := f.section(t, a[0]); got != f.colA {
		t.Errorf("section = %d after rollback, want %d", got, f.colA)
	}
	f.wantOrders(t, []int64{a[0], a[1], f.taskB}, []int{0, 0, 0})
}

func TestMoveTaskToColumnOrderedRejectsBadLists(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	tooMany := make([]int64, domain.MaxOrderedTaskIDs+1)
	for i := range tooMany {
		tooMany[i] = int64(i + 1)
	}
	tooMany[0] = a[0]
	cases := map[string][]int64{
		"duplicate":        {a[0], a[1], a[1]},
		"missing moved id": {a[1], a[2]},
		"zero id":          {a[0], 0},
		"over 500":         tooMany,
	}
	for name, ids := range cases {
		t.Run(name, func(t *testing.T) {
			err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[0], f.colB, testutil.OwnerID, ids)
			if !errors.Is(err, domain.ErrInvalidTaskOrder) {
				t.Fatalf("err = %v, want ErrInvalidTaskOrder", err)
			}
			if got := f.section(t, a[0]); got != f.colA {
				t.Errorf("section changed to %d on a rejected call", got)
			}
			f.wantOrders(t, a, []int{0, 0, 0})
		})
	}
}

func TestMoveTaskToColumnOrderedOnlyBumpsMovedTaskUpdatedAt(t *testing.T) {
	f := newOrderFixture(t)
	a := f.tasksA
	old := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Second)
	if _, err := f.db.ExecContext(f.ctx, "UPDATE flow_tasks SET updated_at = ? WHERE id IN (?, ?, ?)", old, a[0], a[1], a[2]); err != nil {
		t.Fatal(err)
	}
	order := []int64{a[1], a[0], a[2]}
	if err := f.repo.MoveTaskToColumnOrdered(f.ctx, a[0], f.colB, testutil.OwnerID, order); err != nil {
		t.Fatalf("MoveTaskToColumnOrdered: %v", err)
	}
	updated := func(id int64) time.Time {
		var ts time.Time
		if err := f.db.QueryRowContext(f.ctx, "SELECT updated_at FROM flow_tasks WHERE id = ?", id).Scan(&ts); err != nil {
			t.Fatal(err)
		}
		return ts
	}
	if got := updated(a[0]); got.Sub(old) < time.Hour {
		t.Errorf("moved task updated_at = %v, want bumped", got)
	}
	for _, id := range []int64{a[1], a[2]} {
		if got := updated(id); got.Sub(old) > time.Minute || old.Sub(got) > time.Minute {
			t.Errorf("task %d updated_at = %v, want unchanged (%v)", id, got, old)
		}
	}
}
