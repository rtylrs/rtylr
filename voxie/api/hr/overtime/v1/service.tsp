import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.overtime.v1;

/** OvertimeService defines HR overtime request endpoints. */
@route("/hr/overtime/v1")
interface OvertimeService {
  /** ListOvertimeRequests returns overtime requests with optional filters. */
  @get
  @route("/requests")
  listOvertimeRequests(
    ...ListOvertimeRequestsRequest,
  ): ListOvertimeRequestsResponse;

  /** CreateOvertimeRequest submits an overtime request. */
  @post
  @route("/requests")
  createOvertimeRequest(
    ...CreateOvertimeRequestRequest,
  ): CreateOvertimeRequestResponse;

  /** ApproveOvertime approves an overtime request. */
  @patch
  @route("/requests/{overtime_request_id}/approve")
  approveOvertime(...ApproveOvertimeRequest): ApproveOvertimeResponse;
}
