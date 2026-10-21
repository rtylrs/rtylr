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
			}},
		},
	}
}

func TestFinanceReportToolPermissionCheck(t *testing.T) {
	if !financeReportsToolCanRead(financeReportToolUser("finance:reports:view")) {
		t.Fatal("finance:reports:view was rejected")
	}
	if financeReportsToolCanRead(financeReportToolUser("finance:invoices:view")) {
		t.Fatal("a different finance permission was accepted for reports")
	}
}

func TestFinanceReportToolRequiresReportViewPermission(t *testing.T) {
	user := financeReportToolUser("finance:invoices:view")
	if financeReportsToolCanRead(user) {
		t.Fatal("a finance app grant without finance:reports:view authorized report reads")
	}
}

func TestFinanceReportToolsRequireReportViewPermission(t *testing.T) {
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
