import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.presence.v1;

/**
 * Compatibility contract for the Flow presence aggregate until it has a
 * first-class backend service package.
 */
@route("/flow")
interface PresenceService {
  @get
  @route("/presence/today")
  getPresence(...GetPresenceRequest): GetPresenceResponse;
}
