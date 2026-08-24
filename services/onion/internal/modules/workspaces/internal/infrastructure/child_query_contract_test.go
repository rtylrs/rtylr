package infrastructure_test

import (
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaceTaskChildListsHaveServerBounds(t *testing.T) {
	testutil.RequireBoundedListQueries(t, "../../sql/queries/workspaces.sql")
}
