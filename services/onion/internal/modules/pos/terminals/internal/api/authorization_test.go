package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/terminals/internal/application"
	"github.com/voxire/onion/internal/modules/pos/terminals/internal/domain"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"

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

type heartbeatRepo struct{ err error }

type terminalScopeResolver struct{}

func (terminalScopeResolver) ActiveBusinessOwner(_ context.Context, businessID int64) (int64, error) {
	return businessID, nil
}
func (terminalScopeResolver) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

func (r heartbeatRepo) List(context.Context, int64) ([]domain.Terminal, error) { return nil, nil }
func (r heartbeatRepo) Create(context.Context, domain.TerminalInput) (int64, error) {
	return 0, nil
}
func (r heartbeatRepo) Update(context.Context, domain.TerminalUpdateInput) error { return nil }
func (r heartbeatRepo) Heartbeat(context.Context, int64, int64) error            { return r.err }

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

func TestHeartbeatTerminal_ReturnsNotFoundForStaleTerminal(t *testing.T) {
	restoreScopeResolver := tenant.SetScopeResolver(terminalScopeResolver{})
	defer restoreScopeResolver()

	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsAdmin: true,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppPOS, Licensed: true, HasAccess: true,
			}},
		},
	})
	api := NewTerminalAPI(application.NewService(ctx, heartbeatRepo{err: domain.ErrTerminalNotFound}))

	terminalID := int64(55)
	_, err := api.HeartbeatTerminal(ctx, &posterminalv1.HeartbeatTerminalRequest{TerminalId: &terminalID})
	if got := terminalStatus(err); got != http.StatusNotFound {
		t.Fatalf("HeartbeatTerminal status = %d, want %d; error: %v", got, http.StatusNotFound, err)
	}
}
