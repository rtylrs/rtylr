package team

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/team/internal/api"
	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	teamjobs "github.com/voxire/onion/internal/modules/organization/team/internal/jobs"
	"github.com/voxire/onion/internal/modules/organization/team/internal/outbox"
	"github.com/voxire/onion/internal/platform/config"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	hubteamaccessv1 "github.com/voxire/voxie/gen/go/hub/team/access/v1"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
	hubteamonboardingv1 "github.com/voxire/voxie/gen/go/hub/team/onboarding/v1"
	hubteampermissionsv1 "github.com/voxire/voxie/gen/go/hub/team/permissions/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

type NotificationPublisher interface {
	CreateNotification(ctx context.Context, ownerID, businessID, userID int64, title, body, notificationType, link string) error
}

type MemberProvisioner interface {
	EnsureEmployeeForUser(ctx context.Context, ownerID, businessID, userID int64, firstName, lastName, email string) error
	EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error
	DeleteProvisionedMember(ctx context.Context, ownerID, businessID, userID int64) error
}

type TerminatedEmployee struct {
	BusinessID int64
	UserID     int64
}

type TerminatedEmployeeSource interface {
	ListTerminatedEmployees(ctx context.Context) ([]TerminatedEmployee, error)
}

type terminatedEmployeeSourceAdapter struct {
	source TerminatedEmployeeSource
}

func (a terminatedEmployeeSourceAdapter) ListTerminatedEmployees(ctx context.Context) ([]application.TerminatedEmployee, error) {
	rows, err := a.source.ListTerminatedEmployees(ctx)
	if err != nil {
		return nil, err
	}
	employees := make([]application.TerminatedEmployee, 0, len(rows))
	for _, row := range rows {
		employees = append(employees, application.TerminatedEmployee{BusinessID: row.BusinessID, UserID: row.UserID})
	}
	return employees, nil
}

type notificationPublisherAdapter struct{ publisher NotificationPublisher }

func (a notificationPublisherAdapter) CreateNotification(ctx context.Context, ownerID, businessID, userID int64, title, body, notificationType, link string) error {
	return a.publisher.CreateNotification(ctx, ownerID, businessID, userID, title, body, notificationType, link)
}

type emailQueueAdapter struct{ db *sql.DB }

func (a emailQueueAdapter) QueueInvite(ctx context.Context, invite application.InviteEmail) error {
	return emailPkg.EnqueueInvitation(ctx, a.db, invite.To, emailPkg.InvitationData{
		InviterName:  invite.InviterName,
		BusinessName: invite.BusinessName,
		InviteLink:   invite.Link,
		ValidDays:    invite.ValidDays,
	})
}

func (a emailQueueAdapter) QueueAccessRequest(ctx context.Context, toEmail, requesterName, requesterEmail, app string) error {
	appUpper := strings.ToUpper(app)
	subject := requesterName + " requested access to " + appUpper
	return emailPkg.EnqueueTemplate(ctx, a.db, emailPkg.PriorityNormal, toEmail, subject, emailPkg.TemplateAccessRequest, emailPkg.AccessRequestData{
		Subject:        subject,
		RequesterName:  requesterName,
		RequesterEmail: requesterEmail,
		App:            appUpper,
		TeamLink:       config.HubAppURL + "/team",
	})
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(ctx, repository).WithEmailNotifications(emailQueueAdapter{db: db}).WithAuditRecorder(outbox.NewRecorder(db)),
		repository: repository,
	}
}

func (m *Module) GetActiveMemberContact(ctx context.Context, businessID, userID int64) (string, string, bool, error) {
	return m.repository.GetActiveMemberContact(ctx, businessID, userID)
}

func (m *Module) ListActiveMemberUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	return m.repository.ListActiveMemberUserIDs(ctx, businessID)
}

func (m *Module) SetNotificationPublisher(publisher NotificationPublisher) {
	if publisher != nil {
		m.service.WithNotificationPublisher(notificationPublisherAdapter{publisher: publisher})
	}
}

func (m *Module) SetMemberProvisioner(provisioner MemberProvisioner) {
	if provisioner != nil {
		m.service.WithMemberProvisioner(provisioner)
	}
}

func (m *Module) SetTerminatedEmployeeSource(source TerminatedEmployeeSource) {
	if source != nil {
		m.service.WithTerminatedEmployeeSource(terminatedEmployeeSourceAdapter{source: source})
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
