import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.subscription.v1;

@route("/hub/subscription/v1")
interface SubscriptionService {
  @get
  @route("/my")
  getMySubscription(...GetMySubscriptionRequest): GetMySubscriptionResponse;

  @get
  @route("/plans")
  listPlans(...ListPlansRequest): ListPlansResponse;

  @get
  @route("/billing")
  getBillingDetail(...GetBillingDetailRequest): GetBillingDetailResponse;

  @get
  @route("/billing/history")
  getBillingHistory(...GetBillingHistoryRequest): GetBillingHistoryResponse;

  @post
  @route("/my")
  upsertSubscription(...UpsertSubscriptionRequest): UpsertSubscriptionResponse;

  @post
  @route("/checkout")
  createCheckout(...CreateCheckoutRequest): CreateCheckoutResponse;
}
