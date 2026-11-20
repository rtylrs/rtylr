import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.shift.v1;

/** ShiftService defines HR shift and roster endpoints. */
@route("/hr/shift/v1")
interface ShiftService {
  /** Returns shifts with optional filters. */
  @get
  @route("/shifts")
  listShifts(...ListShiftsRequest): ListShiftsResponse;

  /** Creates a new shift. */
  @post
  @route("/shifts")
  createShift(...CreateShiftRequest): CreateShiftResponse;

  /** Updates an existing shift. */
  @put
  @route("/shifts")
  updateShift(...UpdateShiftRequest): UpdateShiftResponse;

  /** Deletes a shift. */
  @delete
  @route("/shifts")
  deleteShift(...DeleteShiftRequest): DeleteShiftResponse;

  /** Assigns an employee to a shift. */
  @post
  @route("/assignments")
  createShiftAssignment(
    ...CreateShiftAssignmentRequest,
  ): CreateShiftAssignmentResponse;

  /** Returns the shift roster for a location and week. */
  @get
  @route("/roster/{location_id}")
  getRoster(...GetRosterRequest): GetRosterResponse;

  /** Publishes the roster for a week. */
  @post
  @route("/roster/publish")
  publishRoster(...PublishRosterRequest): PublishRosterResponse;

  /** Returns staff availability windows with optional filters. */
  @get
  @route("/availability")
  listAvailability(...ListAvailabilityRequest): ListAvailabilityResponse;

  /** Creates or updates one staff availability window. */
  @post
  @route("/availability")
  setAvailability(...SetAvailabilityRequest): SetAvailabilityResponse;

  /** Deletes a staff availability window. */
  @delete
  @route("/availability/{id}")
  deleteAvailability(...DeleteAvailabilityRequest): DeleteAvailabilityResponse;
}
