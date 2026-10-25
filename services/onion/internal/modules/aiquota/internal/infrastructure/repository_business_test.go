package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/platform/tenant"
)

// TestActiveBusinessIDNeverYieldsUsableScope guards the quota fail-closed path:
// a business the platform resolver cannot verify must surface an error and a
// zero id, never an id that would scope quota reads and token-usage writes.
func TestActiveBusinessIDNeverYieldsUsableScope(t *testing.T) {
	for _, tt := range []struct {
		name       string
		businessID int64
		want       error
	}{
		{name: "no business in scope", businessID: 0, want: tenant.ErrNoBusinessScope},
		{name: "negative business id", businessID: -5, want: tenant.ErrNoBusinessScope},
		{name: "resolver unavailable", businessID: 7, want: tenant.ErrOwnerLookupFailed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			businessID, err := NewRepository(nil).ActiveBusinessID(context.Background(), tt.businessID)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ActiveBusinessID error = %v, want %v", err, tt.want)
			}
			if businessID != 0 {
				t.Errorf("ActiveBusinessID = %d, want 0 for an unverified business", businessID)
			}
		})
	}
}
