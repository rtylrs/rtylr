package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
)

type InvitationsAPI struct {
	Service *application.Service
}

func NewInvitationsAPI(service *application.Service) *InvitationsAPI {
	return &InvitationsAPI{Service: service}
}

func (a *InvitationsAPI) InviteUser(ctx context.Context, req *hubteaminvitationsv1.InviteUserRequest) (*hubteaminvitationsv1.InviteUserResponse, error) {
	device, permissionErr := requireHubPermission(ctx, "hub:team:invite")
	if permissionErr != nil {
		return nil, permissionErr
	}

	actorID := device.UserID
	apps := req.GetApps()
	if apps == nil {
		apps = []string{}
	}

	role := domain.Role(req.GetRole())
	if role == "" {
		role = domain.RoleStaff
	}
	appAssignments, assignmentErr := appAssignmentsFromProto(req.GetAppAssignments())
	if assignmentErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
	}
	groupIDs := append([]int64{}, req.GetGroupIds()...)
	password := req.GetPassword()
	provisioningMode := req.GetProvisioningMode()
	if provisioningMode == "direct_credentials" && password == "" {
		password = req.GetTemporaryPassword()
	}

	userID, token, err := a.Service.InviteUser(ctx, actorID, device.BusinessID, domain.InviteUserInput{
		Email:            req.GetEmail(),
		Username:         req.GetUsername(),
		FullName:         req.GetFullName(),
		Password:         password,
		Apps:             apps,
		Role:             role,
		ProvisioningMode: provisioningMode,
		DashRole:         req.GetDashRole(),
		AppAssignments:   appAssignments,
		GroupIDs:         groupIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			return nil, onekiterrors.PermissionDenied()
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "User not found")
		case strings.Contains(strings.ToLower(err.Error()), "invalid role"):
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		case strings.Contains(strings.ToLower(err.Error()), "username") || strings.Contains(strings.ToLower(err.Error()), "direct credentials"):
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to invite user")
		}
	}

	recordSeatChangeAsync(ctx, a.Service, actorID, device.BusinessID)

	if password == "" && token != "" {
		inviteLink := config.AuthAppURL + "/invite/" + token
		if emailErr := a.Service.EnqueueInviteEmail(ctx, req.GetEmail(), inviteLink); emailErr != nil {
			logger.Ctx(ctx).Warn("failed to enqueue invite email", logger.Error(emailErr))
		}
		return &hubteaminvitationsv1.InviteUserResponse{UserId: userID, Message: "Invitation sent"}, nil
	}

	return &hubteaminvitationsv1.InviteUserResponse{UserId: userID, Message: "Member created"}, nil
}

func (a *InvitationsAPI) VerifyInvite(ctx context.Context, req *hubteaminvitationsv1.VerifyInviteRequest) (*hubteaminvitationsv1.VerifyInviteResponse, error) {
	_, email, fullName, _, err := a.Service.VerifyInviteToken(ctx, req.GetToken())
	if err != nil {
		return &hubteaminvitationsv1.VerifyInviteResponse{Valid: false}, nil
	}

	return &hubteaminvitationsv1.VerifyInviteResponse{
		Email:    email,
		FullName: fullName,
		Valid:    true,
	}, nil
}

func (a *InvitationsAPI) AcceptInvite(ctx context.Context, req *hubteaminvitationsv1.AcceptInviteRequest) (*hubteaminvitationsv1.AcceptInviteResponse, error) {
	_, err := a.Service.AcceptInvite(ctx, req.GetToken(), req.GetPassword())
	if err != nil {
		msg := "Invalid or expired invitation"
		if errors.Is(err, domain.ErrTokenAlreadyUsed) {
			msg = "This invitation has already been accepted"
		} else if errors.Is(err, domain.ErrTokenExpired) {
			msg = "This invitation has expired"
		}
		return nil, onekiterrors.New(http.StatusBadRequest, msg)
	}

	return &hubteaminvitationsv1.AcceptInviteResponse{Message: "Password set successfully. You can now sign in."}, nil
}

func (a *InvitationsAPI) BulkInvite(ctx context.Context, req *hubteaminvitationsv1.BulkInviteRequest) (*hubteaminvitationsv1.BulkInviteResponse, error) {
	if _, permissionErr := requireHubPermission(ctx, "hub:team:invite"); permissionErr != nil {
		return nil, permissionErr
	}
	_, ownerID, ok := teamManagerFromCtx(ctx, a.Service)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	emails := req.GetEmails()
	if len(emails) == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "emails is required")
	}

	device := authPkg.DeviceCtx(ctx)
	businessID := int64(0)
	if device.User != nil {
		businessID = device.User.BusinessID
	}

	role := domain.Role(req.GetRole())
	if role == "" {
		role = domain.RoleStaff
	}
	apps := req.GetApps()
	if apps == nil {
		apps = []string{}
	}
	appAssignments, assignmentErr := appAssignmentsFromProto(req.GetAppAssignments())
	if assignmentErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, assignmentErr.Error())
	}
	groupIDs := append([]int64{}, req.GetGroupIds()...)
	if req.GetGroupId() > 0 {
		groupIDs = append(groupIDs, req.GetGroupId())
	}

	results, err := a.Service.BulkInvite(ctx, ownerID, businessID, domain.BulkInviteInput{
		Emails:         emails,
		Role:           role,
		Apps:           apps,
		DashRole:       req.GetDashRole(),
		AppAssignments: appAssignments,
		GroupIDs:       groupIDs,
		GroupID:        req.GetGroupId(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Bulk invite failed")
	}

	payloads := make([]*hubteaminvitationsv1.BulkInviteResult, 0, len(results))
	for _, r := range results {
		payloads = append(payloads, &hubteaminvitationsv1.BulkInviteResult{
			Email:   r.Email,
			UserId:  r.UserID,
			Status:  r.Status,
			Message: r.Message,
		})
	}

	return &hubteaminvitationsv1.BulkInviteResponse{
		Results: payloads,
		Message: "Bulk invite complete",
	}, nil
}
