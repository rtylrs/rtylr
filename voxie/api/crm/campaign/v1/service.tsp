import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.campaign.v1;

/** CampaignService defines CRM campaign endpoints. */
@route("/crm/campaign/v1")
interface CampaignService {
  /** Returns campaigns for the active business. */
  @get
  @route("/campaigns")
  listCampaigns(...ListCampaignsRequest): ListCampaignsResponse;

  /** Creates a new marketing campaign. */
  @post
  @route("/campaigns")
  createCampaign(...CreateCampaignRequest): CreateCampaignResponse;

  /** Returns a campaign and its stats. */
  @get
  @route("/campaigns/{id}")
  getCampaignDetail(...GetCampaignDetailRequest): GetCampaignDetailResponse;

  /** Returns stats for a single campaign. */
  @get
  @route("/campaigns/stats")
  getCampaignStats(...GetCampaignStatsRequest): GetCampaignStatsResponse;

  /** Updates campaign fields. */
  @put
  @route("/campaigns/{id}")
  updateCampaign(...UpdateCampaignRequest): UpdateCampaignResponse;

  /**
   * Changes the campaign lifecycle status. The id is taken from the path;
   * only the status travels in the body (matches PUT /crm/campaigns/{id}/status).
   */
  @put
  @route("/campaigns/{id}/status")
  updateCampaignStatus(
    ...UpdateCampaignStatusRequest,
  ): UpdateCampaignStatusResponse;
}
