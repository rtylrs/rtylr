package api

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/testutil"
)

// A raw DELETE on the users row would remove this console admin; the generic
// resource route has to go through the guarded user deletion instead.
func TestDeleteResourceRowSendsUsersThroughGuardedDelete(t *testing.T) {
	db := testutil.OpenDB(t)
	const adminID int64 = 9_999_947
	cleanup := func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = ?`, adminID) }
	cleanup()
	t.Cleanup(cleanup)
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO users (id, email, password_hash, is_console_admin) VALUES (?, 'delete-route-admin@example.test', 'x', TRUE)`, adminID); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := application.NewService(context.Background(), infrastructure.NewRepository(db))
	ctx := serverHttp.WithPathParams(fix03ConsoleAdminContext(), map[string]string{
		"table": "users",
		"id":    strconv.FormatInt(adminID, 10),
	})
	_, err := (&deleteResourceRowHandler{Service: service}).Serve(ctx, &deleteResourceRowRequest{})
	if got := fix03ConsoleErrorStatus(err); got != http.StatusConflict {
		t.Fatalf("status = %d (%v), want %d", got, err, http.StatusConflict)
	}

	var remaining int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users WHERE id = ?`, adminID).Scan(&remaining); err != nil {
		t.Fatalf("count: %v", err)
	}
	if remaining != 1 {
		t.Error("the guarded delete must leave a console admin in place")
	}
}
