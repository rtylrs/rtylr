import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.campaign.v1;

/** Campaign is a marketing campaign sent to a segment or the full customer base. */
model Campaign {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  description: string;

  /** Delivery channel, e.g. "email" or "sms". */
  channel: string;

  /** Current lifecycle state. */
  status: string;

  @encode(string)
  segment_id?: int64;

  subject: string;
  content: string;
  scheduled_at?: string;
  sent_at?: string;

  @encode(string)
  audience_count: int64;

  @encode(string)
  sent_count: int64;

  @encode(string)
  delivered_count: int64;

  @encode(string)
  opened_count: int64;

  @encode(string)
  clicked_count: int64;

  created_at: string;
  updated_at: string;
}

/** CampaignStats holds delivery and engagement counters for a campaign. */
model CampaignStats {
  @encode(string)
  campaign_id: int64;

  @encode(string)
  audience_count: int64;

  @encode(string)
  sent_count: int64;

  @encode(string)
  delivered_count: int64;

  @encode(string)
  opened_count: int64;

  @encode(string)
  clicked_count: int64;

  @encode(string)
  bounced_count: int64;

  @encode(string)
  failed_count: int64;
}

model CreateCampaignRequest {
  @Onkt.required
  name: string;

  /**
   * Free-text summary. Optional: the server does not require it, and a
   * campaign entered without one is valid.
   */
  description: string;

  @Onkt.required
  channel: string;

  @encode(string)
  segment_id?: int64;

  /**
   * Subject line. Optional because it is channel-specific: an SMS campaign has
   * no subject, so requiring it made that channel impossible to save.
   */
  subject: string;

  @Onkt.required
  content: string;

  /**
   * Absent for a draft or an immediate send. A draft has no scheduled time, so
   * requiring one forced callers to invent a value, and the server already
   * treats an empty scheduled_at as "not scheduled".
   */
  scheduled_at?: string;
}

model CreateCampaignResponse {
  @encode(string)
  campaign_id: int64;

  message: string;
}

model GetCampaignDetailRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetCampaignDetailResponse {
  campaign: Campaign;
  stats: CampaignStats;
}

model GetCampaignStatsRequest {
  /**
   * The live route is GET /campaigns/stats?campaign_id=... (query param,
   * not a path segment).
   */
  @Onkt.required
  @minValueExclusive(0)
  @query("campaign_id")
  @encode(string)
  campaign_id?: int64;
}

model GetCampaignStatsResponse {
  stats: CampaignStats;
}

model ListCampaignsRequest {
  /** Filters campaigns by lifecycle state. */
  @query("status")
  status: string;

  @query("limit")
  @encode(string)
  limit: int64;

  @query("offset")
  @encode(string)
  offset: int64;
}

model ListCampaignsResponse {
  campaigns: Campaign[];

  @encode(string)
  total: int64;

  @encode(string)
  limit: int64;

  @encode(string)
  offset: int64;
}

model UpdateCampaignBody {
  @Onkt.required
  name: string;

  /**
   * Free-text summary. Optional: the server does not require it, and a
   * campaign entered without one is valid.
   */
  description: string;

  @Onkt.required
  channel: string;

  @encode(string)
  segment_id?: int64;

  /**
   * Subject line. Optional because it is channel-specific: an SMS campaign has
   * no subject, so requiring it made that channel impossible to save.
   */
  subject: string;

  @Onkt.required
  content: string;

  /**
   * Absent for a draft or an immediate send. A draft has no scheduled time, so
   * requiring one forced callers to invent a value, and the server already
   * treats an empty scheduled_at as "not scheduled".
   */
  scheduled_at?: string;
}

model UpdateCampaignRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateCampaignBody;
}

model UpdateCampaignResponse {
  message: string;
}

model UpdateCampaignStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateCampaignStatusBody;
}

model UpdateCampaignStatusBody {
  @Onkt.required
  status: string;
}

model UpdateCampaignStatusResponse {
  message: string;
}
