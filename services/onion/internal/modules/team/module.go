package team

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/team/internal/api"
	"github.com/voxire/onion/internal/modules/team/internal/application"
	teamjobs "github.com/voxire/onion/internal/modules/team/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	hubteamaccessv1 "github.com/voxire/voxie/gen/go/hub/team/access/v1"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
	hubteamonboardingv1 "github.com/voxire/voxie/gen/go/hub/team/onboarding/v1"
	hubteampermissionsv1 "github.com/voxire/voxie/gen/go/hub/team/permissions/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubteammembersv1.RegisterMembersServiceServer(
		api.NewMembersAPI(m.service),
		hubteammembersv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team members service: %w", err))
	}

	if err := hubteaminvitationsv1.RegisterInvitationsServiceServer(
		api.NewInvitationsAPI(m.service),
		hubteaminvitationsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team invitations service: %w", err))
	}

	if err := hubteampermissionsv1.RegisterPermissionsServiceServer(
		api.NewPermissionsAPI(m.service),
		hubteampermissionsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team permissions service: %w", err))
	}

	if err := hubteamonboardingv1.RegisterOnboardingServiceServer(
		api.NewOnboardingAPI(m.service),
		hubteamonboardingv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team onboarding service: %w", err))
	}

	if err := hubteamaccessv1.RegisterAccessServiceServer(
		api.NewAccessAPI(m.service),
		hubteamaccessv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team access service: %w", err))
	}

	if err := hubteamgroupsv1.RegisterGroupsServiceServer(
		api.NewGroupsAPI(m.service),
		hubteamgroupsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub team groups service: %w", err))
	}
}

func (m *Module) Service() *application.Service {
	return m.service
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/team/invitations/v1/invitations/verify",
		"/hub/team/invitations/v1/invitations/accept",
	}
}

// RegisterJobs registers the terminated-employee access reconciliation job
// so the worker picks it up via platformjobs.AutoRegister (#834).
func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	teamjobs.Register(ctx, registry, m.service)
}
