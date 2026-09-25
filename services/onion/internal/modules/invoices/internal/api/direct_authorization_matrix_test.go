package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "finance invoices view", App: authorization.AppFinance, Permission: "finance:invoices:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewInvoicesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListInvoices(ctx, nil)
			return response, err
		}},
		{Name: "finance invoices create", App: authorization.AppFinance, Permission: "finance:invoices:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewInvoicesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateInvoice(ctx, &financeinvoicesv1.CreateInvoiceRequest{ClientName: "Authorization Probe", IssueDate: "2026-08-27", DueDate: "2026-09-27", Currency: "USD", Items: []*financeinvoicesv1.InvoiceItem{{Description: "Authorization Probe", Quantity: 1, UnitPrice: 1}}})
			return response, err
		}},
		{Name: "finance invoices send", App: authorization.AppFinance, Permission: "finance:invoices:send", Call: func(ctx context.Context) (any, error) {
			response, err := NewInvoicesAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).SendInvoice(ctx, &financeinvoicesv1.SendInvoiceRequest{Id: 1})
			return response, err
		}},
		{Name: "finance tax configure", App: authorization.AppFinance, Permission: "finance:tax:configure", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListTaxConfigHandler{FinanceService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListTaxConfigRequest{})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
