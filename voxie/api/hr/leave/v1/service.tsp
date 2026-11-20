import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.leave.v1;

/** LeaveService defines HR leave endpoints. */
@route("/hr/leave/v1")
interface LeaveService {
  /** ListLeaveRequests returns leave requests with optional filters. */
  @get
  @route("/requests")
  listLeaveRequests(...ListLeaveRequestsRequest): ListLeaveRequestsResponse;

  /** CreateLeaveRequest submits a leave request. */
  @post
  @route("/requests")
  createLeaveRequest(...CreateLeaveRequestRequest): CreateLeaveRequestResponse;

  /** ApproveLeave approves a leave request. */
  @patch
  @route("/requests/{request_id}/approve")
  approveLeave(...ApproveLeaveRequest): ApproveLeaveResponse;

  /** CancelLeaveRequest withdraws the caller's own pending leave request. */
  @patch
  @route("/requests/{request_id}/cancel")
  cancelLeaveRequest(...CancelLeaveRequestRequest): CancelLeaveResponse;

  /** ListLeaveTypes returns configured leave types. */
  @get
  @route("/types")
  listLeaveTypes(...ListLeaveTypesRequest): ListLeaveTypesResponse;

  /** CreateLeaveType creates a leave type. */
  @post
  @route("/types")
  createLeaveType(...CreateLeaveTypeRequest): CreateLeaveTypeResponse;

  /** UpdateLeaveType updates an existing leave type. */
  @put
  @route("/types/{id}")
  updateLeaveType(...UpdateLeaveTypeRequest): UpdateLeaveTypeResponse;

  /** ListLeaveBalances returns leave balances with optional filters. */
  @get
  @route("/balances")
  listLeaveBalances(...ListLeaveBalancesRequest): ListLeaveBalancesResponse;
}
