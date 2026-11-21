import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.profile.v1;

/**
 * ProfileService exposes business-member profiles with public-project
 * discoverability. Only callers inside the same business can read a profile.
 */
@route("/flow/profile/v1")
interface ProfileService {
  @get
  @route("/profile/{user_id}")
  getProfile(...GetProfileRequest): GetProfileResponse;
}
