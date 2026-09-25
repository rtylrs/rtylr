package api

import (
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

func mappedStatus(t *testing.T, err error) int {
	t.Helper()
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	t.Fatalf("expected a onekit error, got %#v", err)
	return 0
}

// Purging a project that is not archived is a client error: the caller asked
// for something the rules forbid and can fix by archiving first. It used to
// reach the generated server as a bare errors.New and render as a 500 with the
// reason discarded (#1071).
func TestPurgingAnUnarchivedProjectIsAClientError(t *testing.T) {
	mapped := mapProjectV2Error(application.ErrProjectNotArchived)
	if mapped == nil {
		t.Fatal("ErrProjectNotArchived is unmapped, so the caller falls through to a 500")
	}
	if got := mappedStatus(t, mapped); got != http.StatusConflict {
		t.Errorf("status = %d, want %d", got, http.StatusConflict)
	}
	if mapped.Error() == "" {
		t.Error("the mapped error carries no message, so the caller cannot learn to archive first")
	}
}

func TestUnauthorizedStillMapsToForbidden(t *testing.T) {
	mapped := mapProjectV2Error(domain.ErrUnauthorized)
	if got := mappedStatus(t, mapped); got != http.StatusForbidden {
		t.Errorf("status = %d, want %d", got, http.StatusForbidden)
	}
}

// The mapper is called as `if err := mapProjectV2Error(err); err != nil`, so
// returning a non-nil value for an error it does not recognise makes every
// caller's own fallback unreachable. Four handlers depend on that fallback.
func TestUnrecognisedErrorsFallThroughToTheCallersOwnMessage(t *testing.T) {
	if mapped := mapProjectV2Error(errors.New("some repository failure")); mapped != nil {
		t.Errorf("mapProjectV2Error returned %v for an unrecognised error; "+
			"it must return nil so the handler's own fallback fires", mapped)
	}
}
