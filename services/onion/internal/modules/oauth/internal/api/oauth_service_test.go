package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/oauth/internal/application"
	"github.com/voxire/onion/internal/testutil"
	commonoauthv1 "github.com/voxire/voxie/gen/go/common/oauth/v1"
)

// TestRevoke_ReportsFailureInsteadOfSuccess is the B08-R03 regression: a
// revoked refresh token is the whole point of the endpoint, so an
// infrastructure failure must not be answered with Ok:true (RFC 7009 clients
// read the 200 as confirmation). A canceled context makes the repository's
// UPDATE fail deterministically without touching token rows.
func TestRevoke_ReportsFailureInsteadOfSuccess(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	oauthAPI := NewOauthAPI(application.NewService(context.Background(), db))

	resp, err := oauthAPI.Revoke(ctx, &commonoauthv1.RevokeRequest{Token: "refresh-token-that-stays-valid"})
	if err == nil {
		t.Fatal("Revoke reported success while the refresh token update failed")
	}
	if resp != nil {
		t.Errorf("Revoke returned %+v alongside error %v, want no response", resp, err)
	}

	var oneKitErr *onekithttp.Error
	if !errors.As(err, &oneKitErr) {
		t.Fatalf("expected a onekit HTTP error, got %T: %v", err, err)
	}
	if oneKitErr.HTTPStatusCode() != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", oneKitErr.HTTPStatusCode(), http.StatusInternalServerError)
	}
}

func TestRevoke_RejectsEmptyToken(t *testing.T) {
	oauthAPI := &OauthAPI{}

	if _, err := oauthAPI.Revoke(context.Background(), &commonoauthv1.RevokeRequest{}); err == nil {
		t.Fatal("expected an error for an empty token")
	}
}
