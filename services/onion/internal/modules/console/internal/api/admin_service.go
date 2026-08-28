package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	consoleadminv1 "github.com/voxire/voxie/gen/go/console/admin/v1"
)

type ConsoleAdminAPI struct {
	service *application.Service
}

func NewConsoleAdminAPI(service *application.Service) *ConsoleAdminAPI {
	return &ConsoleAdminAPI{service: service}
}

func requireConsoleAdmin(ctx context.Context) error {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	if user.Role != "console_admin" {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func internalError(message string) error {
	return onekiterrors.New(http.StatusInternalServerError, message)
}
func badRequest(message string) error { return onekiterrors.New(http.StatusBadRequest, message) }

func (a *ConsoleAdminAPI) GetStats(ctx context.Context, _ *consoleadminv1.GetStatsRequest) (*consoleadminv1.GetStatsResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	stats, err := a.service.GetStats(ctx)
	if err != nil {
		return nil, internalError("Failed to load console stats")
	}
	return &consoleadminv1.GetStatsResponse{Stats: &consoleadminv1.Stats{
		TotalUsers: stats.TotalUsers, BannedUsers: stats.BannedUsers,
		ActiveSubscriptions: stats.ActiveSubscriptions, TotalBusinesses: stats.TotalBusinesses,
		TotalGrants: stats.TotalGrants, ManualAddonGrants: stats.ManualAddonGrants,
		ConsoleAdmins: stats.ConsoleAdmins,
	}}, nil
}

func (a *ConsoleAdminAPI) ListUsers(ctx context.Context, req *consoleadminv1.ListUsersRequest) (*consoleadminv1.ListUsersResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	limit := int(req.GetLimit())
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	id := int64(0)
	if req.Id != nil {
		id = req.GetId()
	}
	result, err := a.service.ListUsers(ctx, limit, int(req.GetOffset()), id, req.GetEmail(), req.GetName())
	if err != nil {
		return nil, internalError("Failed to list users")
	}
	users := make([]*consoleadminv1.User, 0, len(result.Users))
	for i := range result.Users {
		users = append(users, userPayload(&result.Users[i]))
	}
	return &consoleadminv1.ListUsersResponse{Users: users, Total: result.Total}, nil
}

func (a *ConsoleAdminAPI) GetUser(ctx context.Context, req *consoleadminv1.GetUserRequest) (*consoleadminv1.GetUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	user, err := a.service.GetUser(ctx, req.GetId())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, onekiterrors.New(http.StatusNotFound, "User not found")
	}
	if err != nil {
		return nil, internalError("Failed to get user")
	}
	return &consoleadminv1.GetUserResponse{User: userPayload(user)}, nil
}

func (a *ConsoleAdminAPI) CreateUser(ctx context.Context, req *consoleadminv1.CreateUserRequest) (*consoleadminv1.CreateUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	user, err := a.service.CreateUser(ctx, strings.TrimSpace(req.GetEmail()), req.GetPassword(), strings.TrimSpace(req.GetFullName()), req.GetIsConsoleAdmin())
	if err != nil {
		return nil, internalError("Failed to create user")
	}
	return &consoleadminv1.CreateUserResponse{User: userPayload(user), Message: "User created"}, nil
}

func (a *ConsoleAdminAPI) UpdateUser(ctx context.Context, req *consoleadminv1.UpdateUserRequest) (*consoleadminv1.UpdateUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	current, err := a.service.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, internalError("Failed to get user")
	}
	name, email := current.FullName, current.Email
	if req.FullName != nil {
		name = strings.TrimSpace(req.GetFullName())
	}
	if req.Email != nil {
		email = strings.TrimSpace(req.GetEmail())
	}
	if err := a.service.UpdateUser(ctx, req.GetId(), name, email); err != nil {
		return nil, internalError("Failed to update user")
	}
	return &consoleadminv1.UpdateUserResponse{Message: "User updated"}, nil
}

func (a *ConsoleAdminAPI) DeleteUser(ctx context.Context, req *consoleadminv1.DeleteUserRequest) (*consoleadminv1.DeleteUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.DeleteUser(ctx, req.GetId()); err != nil {
		return nil, internalError("Failed to delete user")
	}
	return &consoleadminv1.DeleteUserResponse{Message: "User deleted"}, nil
}

func (a *ConsoleAdminAPI) BanUser(ctx context.Context, req *consoleadminv1.BanUserRequest) (*consoleadminv1.BanUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.BanUser(ctx, req.GetId(), strings.TrimSpace(req.GetReason())); err != nil {
		return nil, internalError("Failed to ban user")
	}
	return &consoleadminv1.BanUserResponse{Message: "User banned"}, nil
}

func (a *ConsoleAdminAPI) UnbanUser(ctx context.Context, req *consoleadminv1.UnbanUserRequest) (*consoleadminv1.UnbanUserResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.UnbanUser(ctx, req.GetId()); err != nil {
		return nil, internalError("Failed to unban user")
	}
	return &consoleadminv1.UnbanUserResponse{Message: "User unbanned"}, nil
}

func (a *ConsoleAdminAPI) SetConsoleAdmin(ctx context.Context, req *consoleadminv1.SetConsoleAdminRequest) (*consoleadminv1.SetConsoleAdminResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.SetConsoleAdmin(ctx, req.GetId(), req.GetIsConsoleAdmin()); err != nil {
		return nil, internalError("Failed to update console access")
	}
	return &consoleadminv1.SetConsoleAdminResponse{Message: "Console access updated"}, nil
}

func (a *ConsoleAdminAPI) GetUserSubscription(ctx context.Context, req *consoleadminv1.GetUserSubscriptionRequest) (*consoleadminv1.GetUserSubscriptionResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	subscription, err := a.service.GetSubscriptionByOwner(ctx, req.GetUserId())
	if errors.Is(err, sql.ErrNoRows) {
		return &consoleadminv1.GetUserSubscriptionResponse{}, nil
	}
	if err != nil {
		return nil, internalError("Failed to get subscription")
	}
	return &consoleadminv1.GetUserSubscriptionResponse{Subscription: subscriptionPayload(subscription)}, nil
}

func (a *ConsoleAdminAPI) ListBusinesses(ctx context.Context, req *consoleadminv1.ListBusinessesRequest) (*consoleadminv1.ListBusinessesResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	limit := int(req.GetLimit())
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	id, ownerID := int64(0), int64(0)
	if req.Id != nil {
		id = req.GetId()
	}
	if req.OwnerId != nil {
		ownerID = req.GetOwnerId()
	}
	result, err := a.service.ListBusinesses(ctx, limit, int(req.GetOffset()), id, req.GetName(), ownerID)
	if err != nil {
		return nil, internalError("Failed to list businesses")
	}
	businesses := make([]*consoleadminv1.Business, 0, len(result.Businesses))
	for i := range result.Businesses {
		businesses = append(businesses, businessPayload(&result.Businesses[i]))
	}
	return &consoleadminv1.ListBusinessesResponse{Businesses: businesses, Total: result.Total}, nil
}

func (a *ConsoleAdminAPI) UpdateBusiness(ctx context.Context, req *consoleadminv1.UpdateBusinessRequest) (*consoleadminv1.UpdateBusinessResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.UpdateBusiness(ctx, req.GetId(), strings.TrimSpace(req.GetName())); err != nil {
		return nil, internalError("Failed to update business")
	}
	return &consoleadminv1.UpdateBusinessResponse{Message: "Business updated"}, nil
}

func (a *ConsoleAdminAPI) DeleteBusiness(ctx context.Context, req *consoleadminv1.DeleteBusinessRequest) (*consoleadminv1.DeleteBusinessResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.DeleteBusiness(ctx, req.GetId()); err != nil {
		return nil, internalError("Failed to delete business")
	}
	return &consoleadminv1.DeleteBusinessResponse{Message: "Business deleted"}, nil
}

func (a *ConsoleAdminAPI) CreateBusinessOwner(ctx context.Context, req *consoleadminv1.CreateBusinessOwnerRequest) (*consoleadminv1.CreateBusinessOwnerResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := a.service.CreateBusinessOwner(ctx, application.CreateBusinessOwnerInput{
		Email: req.GetEmail(), Password: req.GetPassword(), FullName: req.GetFullName(), BusinessName: req.GetBusinessName(),
		Industry: req.GetIndustry(), Country: req.GetCountry(), Timezone: req.GetTimezone(), Currency: req.GetCurrency(),
		GrantPlan: req.GetGrantPlan(), PlanID: req.GetPlanId(), SubscriptionStatus: req.GetSubscriptionStatus(), GrantDays: int(req.GetGrantDays()),
	})
	if err != nil {
		return nil, badRequest(err.Error())
	}
	response := &consoleadminv1.CreateBusinessOwnerResponse{Owner: userPayload(&result.Owner), Business: businessPayload(&result.Business), Message: result.Message}
	if result.Subscription != nil {
		response.Subscription = subscriptionPayload(result.Subscription)
	}
	return response, nil
}

func (a *ConsoleAdminAPI) ListSubscriptions(ctx context.Context, req *consoleadminv1.ListSubscriptionsRequest) (*consoleadminv1.ListSubscriptionsResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	limit := int(req.GetLimit())
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := a.service.ListSubscriptions(ctx, limit, int(req.GetOffset()), req.GetStatus())
	if err != nil {
		return nil, internalError("Failed to list subscriptions")
	}
	subscriptions := make([]*consoleadminv1.Subscription, 0, len(result.Subscriptions))
	for i := range result.Subscriptions {
		subscriptions = append(subscriptions, subscriptionPayload(&result.Subscriptions[i]))
	}
	return &consoleadminv1.ListSubscriptionsResponse{Subscriptions: subscriptions, Total: result.Total}, nil
}

func (a *ConsoleAdminAPI) UpsertSubscription(ctx context.Context, req *consoleadminv1.UpsertSubscriptionRequest) (*consoleadminv1.UpsertSubscriptionResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	trial, err := parseOptionalTime(req.TrialEndsAt)
	if err != nil {
		return nil, badRequest("Invalid trial_ends_at")
	}
	next, err := parseOptionalTime(req.NextBillingDate)
	if err != nil {
		return nil, badRequest("Invalid next_billing_date")
	}
	period, err := parseOptionalTime(req.CurrentPeriodEndsAt)
	if err != nil {
		return nil, badRequest("Invalid current_period_ends_at")
	}
	var days *int
	if req.TotalDays != nil {
		value := int(req.GetTotalDays())
		days = &value
	}
	err = a.service.UpsertSubscription(ctx, application.SubscriptionWriteInput{
		OwnerID: req.GetOwnerId(), PlanID: req.GetPlanId(), Status: req.GetStatus(), SeatCount: int(req.GetSeatCount()),
		TrialEndsAt: trial, NextBillingDate: next, CurrentPeriodEndsAt: period, TotalDays: days,
		PaddleSubscriptionID: req.GetPaddleSubscriptionId(), PaddleCustomerID: req.GetPaddleCustomerId(),
	})
	if errors.Is(err, application.ErrNotOrganizationOwner) || errors.Is(err, application.ErrInvalidSubscription) {
		return nil, badRequest(err.Error())
	}
	if err != nil {
		return nil, internalError("Failed to upsert subscription")
	}
	return &consoleadminv1.UpsertSubscriptionResponse{Message: "Subscription updated"}, nil
}

func (a *ConsoleAdminAPI) DeleteSubscription(ctx context.Context, req *consoleadminv1.DeleteSubscriptionRequest) (*consoleadminv1.DeleteSubscriptionResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	if err := a.service.DeleteSubscription(ctx, req.GetId()); err != nil {
		return nil, internalError("Failed to delete subscription")
	}
	return &consoleadminv1.DeleteSubscriptionResponse{Message: "Subscription deleted"}, nil
}

func (a *ConsoleAdminAPI) ListPlans(ctx context.Context, _ *consoleadminv1.ListPlansRequest) (*consoleadminv1.ListPlansResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	plans, err := a.service.ListPlans(ctx)
	if err != nil {
		return nil, internalError("Failed to list plans")
	}
	items := make([]*consoleadminv1.Plan, 0, len(plans))
	for i := range plans {
		items = append(items, planPayload(&plans[i]))
	}
	return &consoleadminv1.ListPlansResponse{Plans: items}, nil
}

func (a *ConsoleAdminAPI) UpdatePlan(ctx context.Context, req *consoleadminv1.UpdatePlanRequest) (*consoleadminv1.UpdatePlanResponse, error) {
	if err := requireConsoleAdmin(ctx); err != nil {
		return nil, err
	}
	plans, err := a.service.ListPlans(ctx)
	if err != nil {
		return nil, internalError("Failed to get plan")
	}
	var current *application.PlanDTO
	for i := range plans {
		if plans[i].ID == req.GetId() {
			current = &plans[i]
			break
		}
	}
	if current == nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Plan not found")
	}
	name, monthly, annual, active := current.Name, current.PricePerSeatCents, current.AnnualPricePerSeatCents, current.IsActive
	if req.Name != nil {
		name = req.GetName()
	}
	if req.PricePerSeatCents != nil {
		monthly = int(req.GetPricePerSeatCents())
	}
	if req.AnnualPricePerSeatCents != nil {
		annual = int(req.GetAnnualPricePerSeatCents())
	}
	if req.IsActive != nil {
		active = req.GetIsActive()
	}
	if err := a.service.UpdatePlan(ctx, req.GetId(), name, monthly, annual, active); err != nil {
		return nil, internalError("Failed to update plan")
	}
	return &consoleadminv1.UpdatePlanResponse{Message: "Plan updated"}, nil
}

func userPayload(value *application.UserDTO) *consoleadminv1.User {
	return &consoleadminv1.User{Id: value.ID, Email: value.Email, FullName: value.FullName, BusinessId: value.BusinessID, IsConsoleAdmin: value.IsConsoleAdmin, IsBanned: value.IsBanned, BanReason: value.BanReason, BannedAt: value.BannedAt, CreatedAt: value.CreatedAt}
}

func businessPayload(value *application.BusinessDTO) *consoleadminv1.Business {
	return &consoleadminv1.Business{Id: value.ID, OwnerId: value.OwnerID, Name: value.Name, CreatedAt: value.CreatedAt}
}

func subscriptionPayload(value *application.SubscriptionDTO) *consoleadminv1.Subscription {
	var days *int32
	if value.DaysRemaining != nil {
		v := int32(*value.DaysRemaining) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		days = &v
	}
	updated := value.UpdatedAt
	return &consoleadminv1.Subscription{Id: value.ID, OwnerId: value.OwnerID, OwnerName: value.OwnerName, OwnerEmail: value.OwnerEmail, BusinessId: value.BusinessID, BusinessName: value.BusinessName, PlanId: value.PlanID, PlanName: value.PlanName, Status: value.Status, SeatCount: int32(value.SeatCount), ActiveSeatCount: int32(value.ActiveSeatCount), IncludedSeatLimit: int32(value.IncludedSeatLimit), IncludedSeatCount: int32(value.IncludedSeatCount), BillableSeatCount: int32(value.BillableSeatCount), TrialEndsAt: value.TrialEndsAt, NextBillingDate: value.NextBillingDate, CurrentPeriodEndsAt: value.CurrentPeriodEndsAt, PaddleSubscriptionId: value.PaddleSubscriptionID, PaddleCustomerId: value.PaddleCustomerID, DaysRemaining: days, CreatedAt: value.CreatedAt, UpdatedAt: &updated} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func planPayload(value *application.PlanDTO) *consoleadminv1.Plan {
	return &consoleadminv1.Plan{Id: value.ID, Name: value.Name, PricePerSeatCents: int32(value.PricePerSeatCents), AnnualPricePerSeatCents: int32(value.AnnualPricePerSeatCents), IsActive: value.IsActive, CreatedAt: value.CreatedAt} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(*value)); err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("invalid time")
}
