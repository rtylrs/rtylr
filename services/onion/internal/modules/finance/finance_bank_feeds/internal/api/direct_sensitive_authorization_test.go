package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestDirectSensitiveAuthorization(t *testing.T) {
	txn := domain.BankFeedTransaction{
		BankAccountID:    42,
		ExternalID:       "external-42",
		MatchedExpenseID: int64Pointer(7),
		MatchedInvoiceID: int64Pointer(8),
		OccurredAt:       time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC),
	}

	withoutSensitive := &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppFinance, Licensed: true, HasAccess: true, Permissions: []string{"finance:bank_feeds:view"},
	}}}}
	redacted := toViewerPayload(txn, authPkg.HasAppPermission(withoutSensitive, authorization.AppFinance, "finance:bank_feeds:sensitive_view"))
	if redacted.BankAccountID != 0 || redacted.ExternalID != "" || redacted.MatchedExpenseID != nil || redacted.MatchedInvoiceID != nil {
		t.Fatalf("sensitive bank fields leaked without permission: %+v", redacted)
	}

	withSensitive := &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppFinance, Licensed: true, HasAccess: true, Permissions: []string{"finance:bank_feeds:view", "finance:bank_feeds:sensitive_view"},
	}}}}
	visible := toViewerPayload(txn, authPkg.HasAppPermission(withSensitive, authorization.AppFinance, "finance:bank_feeds:sensitive_view"))
	if visible.BankAccountID != txn.BankAccountID || visible.ExternalID != txn.ExternalID || visible.MatchedExpenseID == nil || visible.MatchedInvoiceID == nil {
		t.Fatalf("sensitive bank fields were not returned with permission: %+v", visible)
	}
}

func int64Pointer(value int64) *int64 { return &value }
