package api

import (
	"context"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	withID := func(ctx context.Context, id string) context.Context {
		rctx := map[string]string{}
		rctx["id"] = (id)
		return serverHttp.WithPathParams(ctx, rctx)
	}

	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hr employees view", App: authorization.AppHR, Permission: "hr:employees:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListEmployeesHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &ListEmployeesRequest{})
			return response, err
		}},
		{Name: "hr employees manage", App: authorization.AppHR, Permission: "hr:employees:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateEmployeeHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &CreateEmployeeRequest{})
			return response, err
		}},
		{Name: "hr roles manage", App: authorization.AppHR, Permission: "hr:roles:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListRolesHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(ctx, &ListRolesRequest{})
			return response, err
		}},
		{Name: "hr employee update", App: authorization.AppHR, Permission: "hr:employees:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&UpdateEmployeeHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(withID(ctx, "4110"), &UpdateEmployeeRequest{})
			return response, err
		}},
		{Name: "hr employee terminate", App: authorization.AppHR, Permission: "hr:employees:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&DeleteEmployeeHandler{Service: application.NewService(ctx, moduleinfra.NewRepository(testutil.DirectAuthorizationDB(ctx)))}).Serve(withID(ctx, "4111"), &DeleteEmployeeRequest{})
			return response, err
		}},
	})
}
