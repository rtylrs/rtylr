import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.activity.v1;

/** ActivityService defines hub activity feed endpoints. */
@route("/hub/activity/v1")
interface ActivityService {
  /** Returns recent activity events. */
  @get
  @route("/activity")
  listActivity(...ListActivityRequest): ListActivityResponse;
}
