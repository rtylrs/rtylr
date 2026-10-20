package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/application"
	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/business"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financebankaccountv1 "github.com/voxire/voxie/gen/go/finance/bank_account/v1"
)

const (
	bankListBusinessID = int64(987654411)
	bankListUserID     = int64(987654412)
)

// The portal's bank page sends empty sort_by and sort_dir values. A business
// with no accounts must get an empty list back, not a server error.
func TestListBankAccounts_PortalQueryOverHTTP(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, bankListBusinessID, bankListUserID)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: bankListBusinessID,
		UserID:     bankListUserID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: bankListBusinessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID: authorization.AppFinance, Licensed: true, HasAccess: true,
				Permissions: []string{"finance:bank:view"},
			}},
		},
	})
	repo := infrastructure.NewRepository(db).WithReportingCurrency(business.NewModule(ctx, db).ReportingCurrency)
	api := NewBankAccountAPI(application.NewService(ctx, repo))
	mux := http.NewServeMux()
	if err := financebankaccountv1.RegisterBankAccountServiceServer(api, financebankaccountv1.WithMux(mux)); err != nil {
		t.Fatalf("register: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/finance/bank-account/v1/bank-accounts?page=1&limit=100&sort_by=&sort_dir=", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d %s", rec.Code, rec.Body.String())
	}
	t.Logf("body: %s", rec.Body.String())
}
