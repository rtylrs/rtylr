import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using crm.customer.v1;

namespace crm.segment.v1;

/** Segment is a saved customer filter used for targeted campaigns. */
model Segment {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  description: string;

  /** JSON filter definition for this segment. */
  filters: string;

  @encode(string)
  customer_count: int64;

  last_calculated_at?: string;
  created_at: string;
  updated_at: string;
}

model CreateSegmentRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  description: string;

  @Onkt.required
  filters: string;
}

model CreateSegmentResponse {
  @encode(string)
  segment_id: int64;

  message: string;
}

model DeleteSegmentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteSegmentResponse {
  message: string;
}

model GetSegmentCustomersRequest {
  @minValueExclusive(0)
  @encode(string)
  segment_id: int64;

  @Onkt.required
  @minValueExclusive(0)
  @query("page")
  @encode(string)
  page?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @query("page_size")
  @encode(string)
  page_size?: int64;
}

model GetSegmentCustomersResponse {
  /**
   * Uses the full crm.customer.v1 Customer (not pos.customer.v1's minimal
   * projection) so consumers can render loyalty tier/points without a
   * second lookup.
   */
  customers: Customer[];

  @encode(string)
  total: int64;

  @encode(string)
  page: int64;

  @encode(string)
  page_size: int64;
}

model GetSegmentDetailRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetSegmentDetailResponse {
  segment: Segment;

  /**
   * Uses the full crm.customer.v1 Customer (not pos.customer.v1's minimal
   * projection) so consumers can render loyalty tier/points without a
   * second lookup.
   */
  customers: Customer[];
}

model ListSegmentsRequest {}

model ListSegmentsResponse {
  segments: Segment[];
}

model UpdateSegmentBody {
  @Onkt.required
  name: string;

  @Onkt.required
  description: string;

  @Onkt.required
  filters: string;
}

model UpdateSegmentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateSegmentBody;
}

model UpdateSegmentResponse {
  message: string;
}
