package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

// directLeadID is the lead seeded by testutil/fixtures/70_direct_authorization.sql.
const directLeadID = 7101

func directAPI(ctx context.Context) *API {
	db := testutil.DirectAuthorizationDB(ctx)
	repository := moduleinfra.NewRepository(db).WithMembershipReader(scopes.NewReader(db).IsBusinessUser)
	return New(application.NewService(ctx, repository))
}

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "crm sales view", App: authorization.AppCRM, Permission: "crm:customers:view", Call: func(ctx context.Context) (any, error) {
			return directAPI(ctx).ListLeads(ctx, &crmsalesv1.ListLeadsRequest{})
		}},
		{Name: "crm sales create", App: authorization.AppCRM, Permission: "crm:customers:create", Call: func(ctx context.Context) (any, error) {
			return directAPI(ctx).CreateLead(ctx, &crmsalesv1.CreateLeadRequest{Lead: &crmsalesv1.LeadFields{CompanyName: "Authorization Probe"}})
		}},
		{Name: "crm sales edit", App: authorization.AppCRM, Permission: "crm:customers:edit", Call: func(ctx context.Context) (any, error) {
			return directAPI(ctx).UpdateLead(ctx, &crmsalesv1.UpdateLeadRequest{Id: directLeadID, Body: &crmsalesv1.UpdateLeadBody{
				Lead: &crmsalesv1.LeadFields{CompanyName: "Authorization Probe"},
			}})
		}},
		{Name: "crm sales delete", App: authorization.AppCRM, Permission: "crm:customers:delete", Call: func(ctx context.Context) (any, error) {
			return directAPI(ctx).DeleteLead(ctx, &crmsalesv1.DeleteLeadRequest{Id: directLeadID})
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
