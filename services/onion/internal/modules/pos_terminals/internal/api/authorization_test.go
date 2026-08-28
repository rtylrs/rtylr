package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	posterminalv1 "github.com/voxire/voxie/gen/go/pos/terminal/v1"
)

func terminalStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestTerminalOperationsRequirePOSTerminalPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	api := &TerminalAPI{}
	cases := []func() error{
		func() error { _, err := api.ListTerminals(ctx, &posterminalv1.ListTerminalsRequest{}); return err },
		func() error { _, err := api.CreateTerminal(ctx, &posterminalv1.CreateTerminalRequest{}); return err },
		func() error { _, err := api.UpdateTerminal(ctx, &posterminalv1.UpdateTerminalRequest{}); return err },
		func() error {
			_, err := api.HeartbeatTerminal(ctx, &posterminalv1.HeartbeatTerminalRequest{})
			return err
		},
	}
	for i, call := range cases {
		if err := call(); terminalStatus(err) != http.StatusForbidden {
			t.Fatalf("case %d returned %v, want 403 before service access", i, err)
		}
	}
}
