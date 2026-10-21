package api

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/financejournal"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	withID := func(ctx context.Context) context.Context {
		rctx := map[string]string{}
		rctx["id"] = ("4301")
		return serverHttp.WithPathParams(ctx, rctx)
	}
	newService := func(ctx context.Context) *application.Service {
		service := application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)), nil, hrcore.NewModule(ctx, testutil.DirectAuthorizationDB(ctx)))
		service.SetFinancePoster(directPayrollPoster{})
		return service
	}

	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hr payroll view", App: authorization.AppHR, Permission: "hr:payroll:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListPayrollRunsHandler{Service: newService(ctx)}).Serve(ctx, &ListPayrollRunsRequest{})
			return response, err
		}},
		{Name: "hr payroll manage", App: authorization.AppHR, Permission: "hr:payroll:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&UpsertPayrollConfigHandler{Service: newService(ctx)}).Serve(ctx, &UpsertPayrollConfigRequest{})
			return response, err
		}},
		{Name: "hr payroll execute", App: authorization.AppHR, Permission: "hr:payroll:execute", Call: func(ctx context.Context) (any, error) {
			response, err := (&TriggerPayrollRunHandler{Service: newService(ctx)}).Serve(ctx, &TriggerPayrollRunRequest{PeriodStart: "2027-09-01", PeriodEnd: "2027-09-30"})
			return response, err
		}},
		{Name: "hr payroll export boundary", App: authorization.AppHR, Permission: "hr:payroll:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&ExportPayrollHandler{Service: newService(ctx)}).Serve(withID(ctx), &ExportPayrollRequest{})
			return response, err
		}},
	})
}

type directPayrollPoster struct{}

func (directPayrollPoster) PostPayrollJournal(context.Context, financejournal.PayrollJournalEntry) error {
	return nil
}
