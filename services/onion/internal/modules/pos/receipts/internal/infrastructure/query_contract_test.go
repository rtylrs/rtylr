package infrastructure_test

import (
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

func TestReceiptListQueriesHaveServerBounds(t *testing.T) {
	testutil.RequireBoundedListQueries(t, "../../sql/queries/pos_receipts.sql")
}
