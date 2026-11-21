import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace console.admin.v1;

@route("/console/admin/v1")
interface ConsoleAdminService {
  @get
  @route("/console/stats")
  getStats(...GetStatsRequest): GetStatsResponse;

  @get
  @route("/console/users")
  listUsers(...ListUsersRequest): ListUsersResponse;

  @get
  @route("/console/users/{id}")
  getUser(...GetUserRequest): GetUserResponse;

  @post
  @route("/console/users")
  createUser(...CreateUserRequest): CreateUserResponse;

  @put
  @route("/console/users/{id}")
  updateUser(...UpdateUserRequest): UpdateUserResponse;

  /**
   * Deletes a user's account, access and personal data, and deactivates the
   * businesses only they use. Returns 409 while another member uses a business
   * they own, while their Paddle subscription can still bill, for a console
   * admin, or for the caller's own account.
   */
  @delete
  @route("/console/users/{id}")
  deleteUser(...DeleteUserRequest): DeleteUserResponse;

  @post
  @route("/console/users/{id}/ban")
  banUser(...BanUserRequest): BanUserResponse;

  @post
  @route("/console/users/{id}/unban")
  unbanUser(...UnbanUserRequest): UnbanUserResponse;

  @post
  @route("/console/users/{id}/set-console-admin")
  setConsoleAdmin(...SetConsoleAdminRequest): SetConsoleAdminResponse;

  @get
  @route("/console/users/{user_id}/subscription")
  getUserSubscription(
    ...GetUserSubscriptionRequest,
  ): GetUserSubscriptionResponse;

  @get
  @route("/console/businesses")
  listBusinesses(...ListBusinessesRequest): ListBusinessesResponse;

  @put
  @route("/console/businesses/{id}")
  updateBusiness(...UpdateBusinessRequest): UpdateBusinessResponse;

  @delete
  @route("/console/businesses/{id}")
  deleteBusiness(...DeleteBusinessRequest): DeleteBusinessResponse;

  @post
  @route("/console/internal-onboarding/business-owner")
  createBusinessOwner(
    ...CreateBusinessOwnerRequest,
  ): CreateBusinessOwnerResponse;

  @get
  @route("/console/subscriptions")
  listSubscriptions(...ListSubscriptionsRequest): ListSubscriptionsResponse;

  @post
  @route("/console/subscriptions/upsert")
  upsertSubscription(...UpsertSubscriptionRequest): UpsertSubscriptionResponse;

  @delete
  @route("/console/subscriptions/{id}")
  deleteSubscription(...DeleteSubscriptionRequest): DeleteSubscriptionResponse;

  @get
  @route("/console/plans")
  listPlans(...ListPlansRequest): ListPlansResponse;

  @put
  @route("/console/plans/{id}")
  updatePlan(...UpdatePlanRequest): UpdatePlanResponse;

  @get
  @route("/console/plan-region-prices")
  listPlanRegionPrices(
    ...ListPlanRegionPricesRequest,
  ): ListPlanRegionPricesResponse;

  @put
  @route("/console/plans/{plan_id}/region-prices/{region}")
  upsertPlanRegionPrice(
    ...UpsertPlanRegionPriceRequest,
  ): UpsertPlanRegionPriceResponse;

  @delete
  @route("/console/plans/{plan_id}/region-prices/{region}")
  deletePlanRegionPrice(
    ...DeletePlanRegionPriceRequest,
  ): DeletePlanRegionPriceResponse;
}
