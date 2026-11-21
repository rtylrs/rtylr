import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.loyalty.v1;

/** LoyaltyService defines CRM loyalty program endpoints. */
@route("/crm/loyalty/v1")
interface LoyaltyService {
  /** Returns the loyalty program configuration. */
  @get
  @route("/config")
  getLoyaltyConfig(...GetLoyaltyConfigRequest): GetLoyaltyConfigResponse;

  /** Updates the loyalty program configuration. */
  @patch
  @route("/config")
  updateLoyaltyConfig(
    ...UpdateLoyaltyConfigRequest,
  ): UpdateLoyaltyConfigResponse;

  /** Returns all loyalty tiers. */
  @get
  @route("/tiers")
  listLoyaltyTiers(...ListLoyaltyTiersRequest): ListLoyaltyTiersResponse;

  /** Creates a new loyalty tier. */
  @post
  @route("/tiers")
  createLoyaltyTier(...CreateLoyaltyTierRequest): CreateLoyaltyTierResponse;

  /** Updates an existing loyalty tier. */
  @put
  @route("/tiers/{id}")
  updateLoyaltyTier(...UpdateLoyaltyTierRequest): UpdateLoyaltyTierResponse;

  /** Returns paginated loyalty transactions. */
  @get
  @route("/transactions")
  listLoyaltyTransactions(
    ...ListLoyaltyTransactionsRequest,
  ): ListLoyaltyTransactionsResponse;

  /** Awards loyalty points to a customer. */
  @post
  @route("/award")
  awardPoints(...AwardPointsRequest): AwardPointsResponse;

  /** Redeems loyalty points for a customer. */
  @post
  @route("/redeem")
  redeemPoints(...RedeemPointsRequest): RedeemPointsResponse;
}
