import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.segment.v1;

/** SegmentService defines CRM segment endpoints. */
@route("/crm/segment/v1")
interface SegmentService {
  /** Returns all segments for the active business. */
  @get
  @route("/segments")
  listSegments(...ListSegmentsRequest): ListSegmentsResponse;

  /** Creates a new customer segment. */
  @post
  @route("/segments")
  createSegment(...CreateSegmentRequest): CreateSegmentResponse;

  /** Returns a segment and its member customers. */
  @get
  @route("/segments/{id}")
  getSegmentDetail(...GetSegmentDetailRequest): GetSegmentDetailResponse;

  /** Returns paginated customers in a segment. */
  @get
  @route("/segments/{segment_id}/customers")
  getSegmentCustomers(
    ...GetSegmentCustomersRequest,
  ): GetSegmentCustomersResponse;

  /** Updates an existing segment. */
  @put
  @route("/segments/{id}")
  updateSegment(...UpdateSegmentRequest): UpdateSegmentResponse;

  /** Removes a segment by ID. */
  @delete
  @route("/segments/{id}")
  deleteSegment(...DeleteSegmentRequest): DeleteSegmentResponse;
}
