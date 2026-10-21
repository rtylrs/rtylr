package infrastructure_test

import (
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

func TestGroupListQueriesHaveServerBounds(t *testing.T) {
	testutil.RequireBoundedListQueries(t, "../../sql/queries/groups.sql")
}
