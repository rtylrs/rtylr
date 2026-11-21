import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.compat.v1;

model EmptyRequest {}

/** EmptyResponse marks a method whose handler writes its own response. */
model EmptyResponse {}

model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}

model PageRequest {
  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;
}

model CustomerListRequest {
  @query("search")
  search?: string;

  @query("tier")
  tier?: string;

  @query("tag")
  tag?: string;

  @query("page")
  page?: int32;

  @query("page_size")
  page_size?: int32;

  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;
}

model IdQueryRequest {
  @query("id")
  id?: int32;
}

model CustomerTimelineRequest {
  @query("id")
  id?: int32;

  @query("customer_id")
  customer_id?: int32;

  @query("cursor")
  cursor?: string;

  @query("limit")
  limit?: int32;
}

model SegmentCustomersRequest {
  @query("segment_id")
  segment_id?: int32;
}

model CampaignStatsRequest {
  @query("campaign_id")
  campaign_id?: int32;
}

/**
 * CampaignAudienceRequest counts an email campaign's reach; without a
 * segment the audience is every customer.
 */
model CampaignAudienceRequest {
  @query("segment_id")
  @encode(string)
  segment_id?: int64;
}

model LoyaltyTransactionsRequest {
  @query("customer_id")
  customer_id?: int32;

  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;
}

model ExportCustomerRequest {
  @query("id")
  id?: int32;

  @query("customer_id")
  customer_id?: int32;
}

model BodyRequest {
  @body
  body: unknown;
}

model PathRequest {
  id: string;
}

model PathBodyRequest {
  id: string;

  @body
  body: unknown;
}
