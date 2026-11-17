package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// A reorder that does not list the project's exact set is the caller's mistake
// (usually a stale board), so it has to reach them as a 400 they can act on,
// not a 500 that hides the reason.
func TestStatusOrderMismatchIsABadRequest(t *testing.T) {
	mapped := statusError(fmt.Errorf("reorder: %w", domain.ErrStatusOrderMismatch), "fallback")
	if got := mappedStatus(t, mapped); got != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", got, http.StatusBadRequest)
	}
	if mapped.Error() == "fallback" {
		t.Error("the mismatch fell through to the fallback message")
	}
}
