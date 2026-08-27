package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onkithttp "github.com/corezio/onekit/http"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func automationUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppCRM,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func automationPermissionStatus(err error) int {
	var oneKitErr *onkithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestAutomationHandlersRequireCanonicalPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), automationUser())

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "list",
			call: func() error {
				_, err := (&ListAutomationsHandler{}).Serve(ctx, &ListAutomationsRequest{})
				return err
			},
		},
		{
			name: "create",
			call: func() error {
				_, err := (&CreateAutomationHandler{}).Serve(ctx, &CreateAutomationRequest{})
				return err
			},
		},
		{
			name: "update",
			call: func() error {
				_, err := (&UpdateAutomationHandler{}).Serve(ctx, &UpdateAutomationRequest{})
				return err
			},
		},
		{
			name: "delete",
			call: func() error {
				_, err := (&DeleteAutomationHandler{}).Serve(ctx, &DeleteAutomationRequest{})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := automationPermissionStatus(tc.call()); got != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
			}
		})
	}
}

func TestRequireAutomationManagerChecksCRMAppPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), automationUser("crm:automations:manage"))
	if err := requireAutomationManager(ctx); err != nil {
		t.Fatalf("granted automation permission rejected: %v", err)
	}

	wrongApp := automationUser("crm:automations:manage")
	wrongApp.EffectiveAccess.Apps[0].AppID = authorization.AppFinance
	if got := automationPermissionStatus(requireAutomationManager(authPkg.WithAuthUser(context.Background(), wrongApp))); got != http.StatusForbidden {
		t.Fatalf("wrong app status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestCreateAutomationHandler_Validate_MissingName(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "", TriggerType: "customer_signup"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateAutomationHandler_Validate_MissingTriggerType(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing trigger_type")
	}
}

func TestCreateAutomationHandler_Validate_Valid(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: "customer_signup", Actions: []byte(`[{"type":"send_email"}]`)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateAutomationHandler_Validate_RejectsUnsupportedContract(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: "New customer", Actions: []byte(`[{"type":"send_email"}]`)}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected unsupported trigger error")
	}
}

func TestUpdateAutomationHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestUpdateAutomationHandler_Validate_Valid(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{Name: "Updated Automation", TriggerType: "customer_signup"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateAutomationHandler_Validate_RejectsUnsupportedAction(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{
		Name:        "Updated Automation",
		TriggerType: "customer_signup",
		Actions:     []byte(`[{"type":"send_push_notification"}]`),
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected unsupported action error")
	}
}
