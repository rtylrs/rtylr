import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.access.v1;

/** AccessService defines POS location access endpoints. */
@route("/pos/access/v1")
interface AccessService {
  /** Returns the locations assigned to the current staff member. */
  @get
  @route("/my-locations")
  getMyLocations(...GetMyLocationsRequest): GetMyLocationsResponse;

  /** Returns the locations assigned to a given staff member. */
  @get
  @route("/member-locations")
  getMemberLocations(...GetMemberLocationsRequest): GetMemberLocationsResponse;

  /** Returns the staff roster for a given location. */
  @get
  @route("/location-roster")
  getLocationRoster(...GetLocationRosterRequest): GetLocationRosterResponse;

  /** Assigns a staff member to a location. */
  @post
  @route("/assign")
  assignLocation(...AssignLocationRequest): AssignLocationResponse;

  /** Removes a staff member's location assignment. */
  @post
  @route("/remove")
  removeLocation(...RemoveLocationRequest): RemoveLocationResponse;
}
