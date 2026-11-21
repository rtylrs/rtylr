import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.compat.v1;

/**
 * Typed OneKit compatibility surface for CRM routes that predate the
 * canonical crm.* service packages. It keeps their existing wire paths while
 * removing Axios/safe from app code.
 */
@route("/crm")
interface CompatibilityService {
  @get
  @route("/dashboard")
  getDashboard(...EmptyRequest): JsonResponse;

  @get
  @route("/customers")
  listCustomers(...CustomerListRequest): JsonResponse;

  @get
  @route("/customers/detail")
  getCustomer(...IdQueryRequest): JsonResponse;

  @get
  @route("/customers/detail/full")
  getCustomerFull(...IdQueryRequest): JsonResponse;

  @get
  @route("/customers/timeline")
  getCustomerTimeline(...CustomerTimelineRequest): JsonResponse;

  @post
  @route("/customers")
  createCustomer(...BodyRequest): JsonResponse;

  @put
  @route("/customers")
  updateCustomer(...BodyRequest): JsonResponse;

  @post
  @route("/customers/record-visit")
  recordVisit(...BodyRequest): JsonResponse;

  @post
  @route("/customers/import")
  importCustomers(...BodyRequest): JsonResponse;

  @post
  @route("/customers/merge")
  mergeCustomers(...BodyRequest): JsonResponse;

  @get
  @route("/customers/export")
  exportCustomer(...ExportCustomerRequest): JsonResponse;

  @delete
  @route("/customers")
  deleteCustomer(...IdQueryRequest): JsonResponse;

  @post
  @route("/loyalty/award")
  awardPoints(...BodyRequest): JsonResponse;

  @post
  @route("/loyalty/redeem")
  redeemPoints(...BodyRequest): JsonResponse;

  @get
  @route("/loyalty/config")
  getLoyaltyConfig(...EmptyRequest): JsonResponse;

  @put
  @route("/loyalty/config")
  updateLoyaltyConfig(...BodyRequest): JsonResponse;

  @get
  @route("/loyalty/transactions")
  getLoyaltyTransactions(...LoyaltyTransactionsRequest): JsonResponse;

  @get
  @route("/loyalty/members")
  getLoyaltyMembers(...PageRequest): JsonResponse;

  @get
  @route("/loyalty/rewards")
  getLoyaltyRewards(...EmptyRequest): JsonResponse;

  @post
  @route("/loyalty/rewards/upsert")
  upsertLoyaltyReward(...BodyRequest): JsonResponse;

  @get
  @route("/loyalty/tiers")
  getLoyaltyTiers(...EmptyRequest): JsonResponse;

  @post
  @route("/loyalty/tiers")
  createLoyaltyTier(...BodyRequest): JsonResponse;

  @put
  @route("/loyalty/tiers")
  updateLoyaltyTier(...BodyRequest): JsonResponse;

  @delete
  @route("/loyalty/tiers/{id}")
  deleteLoyaltyTier(...PathRequest): JsonResponse;

  @get
  @route("/segments")
  listSegments(...EmptyRequest): JsonResponse;

  @get
  @route("/segments/detail")
  getSegment(...IdQueryRequest): JsonResponse;

  @get
  @route("/segments/customers")
  getSegmentCustomers(...SegmentCustomersRequest): JsonResponse;

  @post
  @route("/segments")
  createSegment(...BodyRequest): JsonResponse;

  @put
  @route("/segments")
  updateSegment(...BodyRequest): JsonResponse;

  @delete
  @route("/segments")
  deleteSegment(...IdQueryRequest): JsonResponse;

  @get
  @route("/campaigns")
  listCampaigns(...PageRequest): JsonResponse;

  @get
  @route("/campaigns/detail")
  getCampaign(...IdQueryRequest): JsonResponse;

  @get
  @route("/campaigns/{id}")
  getCampaignById(...PathRequest): JsonResponse;

  @get
  @route("/campaigns/stats")
  getCampaignStats(...CampaignStatsRequest): JsonResponse;

  @post
  @route("/campaigns")
  createCampaign(...BodyRequest): JsonResponse;

  @put
  @route("/campaigns/{id}")
  updateCampaign(...PathBodyRequest): JsonResponse;

  @put
  @route("/campaigns/{id}/status")
  updateCampaignStatus(...PathBodyRequest): JsonResponse;

  @delete
  @route("/campaigns/{id}")
  deleteCampaign(...PathRequest): JsonResponse;

  @get
  @route("/campaigns/audience")
  getCampaignAudience(...CampaignAudienceRequest): JsonResponse;

  @post
  @route("/campaigns/{id}/send")
  sendCampaign(...PathRequest): JsonResponse;

  /** The unsubscribe page linked from marketing email; it writes HTML. */
  @get
  @route("/public/unsubscribe")
  @Onkt.http
  getUnsubscribe(...EmptyRequest): EmptyResponse;

  @post
  @route("/public/unsubscribe")
  @Onkt.http
  confirmUnsubscribe(...EmptyRequest): EmptyResponse;

  @get
  @route("/automations")
  listAutomations(...EmptyRequest): JsonResponse;

  @post
  @route("/automations")
  createAutomation(...BodyRequest): JsonResponse;

  @put
  @route("/automations/{id}")
  updateAutomation(...PathBodyRequest): JsonResponse;

  @delete
  @route("/automations/{id}")
  deleteAutomation(...PathRequest): JsonResponse;
}
