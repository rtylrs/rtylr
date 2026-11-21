import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.location.v1;

/** LocationService defines POS location endpoints. */
@route("/pos/location/v1")
interface LocationService {
  /** Returns locations for the active business. */
  @get
  @route("/locations")
  listLocations(...ListLocationsRequest): ListLocationsResponse;

  /** Creates a new location. */
  @post
  @route("/locations")
  createLocation(...CreateLocationRequest): CreateLocationResponse;

  /** Updates an existing location. */
  @put
  @route("/locations/{id}")
  updateLocation(...UpdateLocationRequest): UpdateLocationResponse;
}
