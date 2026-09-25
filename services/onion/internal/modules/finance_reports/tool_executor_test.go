package financereports

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func financeReportToolUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     9101,
		BusinessID: 9201,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:           9201,
			AuthorizationVersion: 1,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFinance,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
				Scopes:      []authorization.Scope{{Type: "finance", ID: "9201"}},
			}},
		},
	}
}

func TestFinanceReportToolPermissionCheck(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	if !financeReportsToolCanRead(financeReportToolUser("finance:reports:view")) {
		t.Fatal("finance:reports:view was rejected")
	}
	if financeReportsToolCanRead(financeReportToolUser("finance:invoices:view")) {
		t.Fatal("a different finance permission was accepted for reports")
	}
}

func TestFinanceReportToolDoesNotUseLegacyAppMembershipFallback(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "legacy")

	user := financeReportToolUser("finance:invoices:view")
	user.Role = "member"
	if financeReportsToolCanRead(user) {
		t.Fatal("legacy finance app membership unexpectedly authorized report reads")
	}
}

func TestFinanceReportToolsRequireReportViewPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	for _, toolName := range []string{getFinanceBalanceSheetToolName, getFinanceCashFlowReportToolName} {
		t.Run(toolName, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), financeReportToolUser("finance:invoices:view"))
			output, handled, err := (&Module{}).ExecuteCaretTool(ctx, toolName, nil)
			if !handled {
				t.Fatal("tool was not recognized")
			}
			if err == nil {
				t.Fatal("report tool without finance:reports:view unexpectedly succeeded")
			}
			if output != nil {
				t.Fatalf("denied report tool returned output: %#v", output)
			}
		})
	}
}
