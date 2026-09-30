package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/tenant"
)

// TestResolveOwnerIDNeverSubstitutesFallbackUser is the regression test for the
// log-and-continue fail-open: this method used to log a warning and return
// fallbackUserID when the platform resolver could not verify an owner, handing
// callers an unrelated user id that then scoped quota reads and token-usage
// writes. It must surface the error and a non-usable id instead.
func TestResolveOwnerIDNeverSubstitutesFallbackUser(t *testing.T) {
	const fallbackUserID = int64(4242)

	tests := []struct {
		name       string
		businessID int64
		wantErr    error
		needNilDB  bool
	}{
		{
			name:       "no business in scope",
			businessID: 0,
			wantErr:    tenant.ErrNoBusinessScope,
		},
		{
			name:       "negative business id",
			businessID: -5,
			wantErr:    tenant.ErrNoBusinessScope,
		},
		{
			// Exercises the same log-and-continue branch as a live failure, but
			// without a database: an absent pool is one of the ways the platform
			// resolver refuses to answer.
			name:       "database unavailable",
			businessID: 7,
			wantErr:    tenant.ErrOwnerLookupFailed,
			needNilDB:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.needNilDB && dbclient.Pool() != nil {
				t.Skip("a database pool is configured; this case needs the resolver to fail closed without one")
			}

			ownerID, err := NewRepository(nil).ResolveOwnerID(context.Background(), tt.businessID, fallbackUserID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ResolveOwnerID error = %v, want %v", err, tt.wantErr)
			}
			if ownerID != 0 {
				t.Errorf("ResolveOwnerID owner id = %d, want 0: a failed resolution must not yield a usable owner (fallback was %d)", ownerID, fallbackUserID)
			}
		})
	}
}
