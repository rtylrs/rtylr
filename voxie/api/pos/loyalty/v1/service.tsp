import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.loyalty.v1;

/** LoyaltyService defines POS loyalty program endpoints. */
@route("/pos/loyalty/v1")
interface LoyaltyService {
  /** Returns the loyalty program configuration for the active business. */
  @get
  @route("/config")
  getLoyaltyConfig(...GetLoyaltyConfigRequest): GetLoyaltyConfigResponse;

  /** Updates the loyalty program configuration. */
  @put
  @route("/config")
  updateLoyaltyConfig(
    ...UpdateLoyaltyConfigRequest,
  ): UpdateLoyaltyConfigResponse;

  /** Returns loyalty transactions, optionally filtered by customer. */
  @get
  @route("/transactions")
  listLoyaltyTransactions(
    ...ListLoyaltyTransactionsRequest,
  ): ListLoyaltyTransactionsResponse;

  /** Records points earned by a customer for an order. */
  @post
  @route("/earn")
  earnPoints(...EarnPointsRequest): EarnPointsResponse;

  /** Redeems loyalty points for a customer. */
  @post
  @route("/redeem")
  redeemPoints(...RedeemPointsRequest): RedeemPointsResponse;

  /** Returns the loyalty points balance for a customer. */
  @get
  @route("/balance")
  getLoyaltyBalance(...GetLoyaltyBalanceRequest): GetLoyaltyBalanceResponse;
}
