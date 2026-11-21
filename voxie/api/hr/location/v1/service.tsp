import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.location.v1;

/** LocationService defines HR location endpoints. */
@route("/hr/location/v1")
interface LocationService {
  /** CreateLocation creates a new location. */
  @post
  @route("/locations")
  createLocation(...CreateLocationRequest): CreateLocationResponse;

  /** UpdateLocation updates an existing location. */
  @put
  @route("/locations/{id}")
  updateLocation(...UpdateLocationRequest): UpdateLocationResponse;
}
