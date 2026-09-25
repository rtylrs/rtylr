package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	consoleadminv1 "github.com/voxire/voxie/gen/go/console/admin/v1"
)

func fix03ConsoleErrorStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func fix03ConsoleAdminContext() context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID: 7,
		Role:   "console_admin",
	})
}

func TestGrantExpiryParsingAcceptsDatePickerValues(t *testing.T) {
	for _, value := range []string{"2026-12-31", "2026-12-31T00:00", "2026-12-31T00:00:00Z"} {
		parsed, err := parseGrantExpiresAt(value)
		if err != nil {
			t.Errorf("parseGrantExpiresAt(%q): %v", value, err)
		} else if parsed == nil {
			t.Errorf("parseGrantExpiresAt(%q) returned nil", value)
		}
	}
}

func TestGrantHandlersRejectMalformedExpiresAt(t *testing.T) {
	ctx := fix03ConsoleAdminContext()
	handlers := []struct {
		name     string
		validate func() error
		serve    func() error
	}{
		{
			name: "access grant",
			validate: func() error {
				return (&CreateGrantHandler{}).Validate(ctx, &CreateGrantRequest{
					UserID: 1, PlanID: 2, ExpiresAt: "2026-13-01T00:00:00Z",
				})
			},
			serve: func() error {
				_, err := (&CreateGrantHandler{}).Serve(ctx, &CreateGrantRequest{
					UserID: 1, PlanID: 2, ExpiresAt: "2026-13-01T00:00:00Z",
				})
				return err
			},
		},
		{
			name: "addon grant",
			validate: func() error {
				return (&CreateAddonGrantHandler{}).Validate(ctx, &CreateAddonGrantRequest{
					UserID: 1, AddonID: "menu", ExpiresAt: "2026-13-01T00:00:00Z",
				})
			},
			serve: func() error {
				_, err := (&CreateAddonGrantHandler{}).Serve(ctx, &CreateAddonGrantRequest{
					UserID: 1, AddonID: "menu", ExpiresAt: "2026-13-01T00:00:00Z",
				})
				return err
			},
		},
	}

	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			if got := fix03ConsoleErrorStatus(tc.validate()); got != http.StatusBadRequest {
				t.Errorf("Validate status = %d, want %d", got, http.StatusBadRequest)
			}
			if got := fix03ConsoleErrorStatus(tc.serve()); got != http.StatusBadRequest {
				t.Errorf("Serve status = %d, want %d", got, http.StatusBadRequest)
			}
		})
	}
}

func TestUpdateBusinessRejectsMissingOrWhitespaceName(t *testing.T) {
	ctx := fix03ConsoleAdminContext()
	api := &ConsoleAdminAPI{}
	requests := []*consoleadminv1.UpdateBusinessRequest{
		{Id: 1},
		{Id: 1, Name: "   "},
	}

	for _, req := range requests {
		_, err := api.UpdateBusiness(ctx, req)
		if got := fix03ConsoleErrorStatus(err); got != http.StatusBadRequest {
			t.Errorf("UpdateBusiness(%q) status = %d, want %d", req.GetName(), got, http.StatusBadRequest)
		}
	}
}
