import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.staff.v1;

/** StaffService defines HR staff endpoints. */
@route("/hr/staff/v1")
interface StaffService {
  /** Returns staff members with optional filters. */
  @get
  @route("/staff")
  listStaff(...ListStaffRequest): ListStaffResponse;

  /** Creates a new staff member. */
  @post
  @route("/staff")
  createStaff(...CreateStaffRequest): CreateStaffResponse;

  /** Returns a staff member by ID. */
  @get
  @route("/staff/{id}")
  getStaff(...GetStaffRequest): GetStaffResponse;

  /** Updates an existing staff member. */
  @put
  @route("/staff/{id}")
  updateStaff(...UpdateStaffRequest): UpdateStaffResponse;

  /** Removes a staff member by ID. */
  @delete
  @route("/staff/{id}")
  deleteStaff(...DeleteStaffRequest): DeleteStaffResponse;

  /** Records a staff clock-in event. */
  @post
  @route("/staff/clock-in")
  clockIn(...ClockInRequest): ClockInResponse;

  /** Records a staff clock-out event. */
  @post
  @route("/staff/clock-out")
  clockOut(...ClockOutRequest): ClockOutResponse;

  /**
   * Returns the paginated, business-wide staff directory. Gated by both
   * the business's directory_enabled config flag and the caller's
   * view_directory permission; distinct from listStaff, which serves
   * HR's own admin screens with the full Staff shape.
   */
  @get
  @route("/staff/directory")
  listStaffDirectory(...ListStaffDirectoryRequest): ListStaffDirectoryResponse;
}
