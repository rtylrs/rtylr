import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.onboarding.v1;

/** OnboardingService defines hub team onboarding endpoints. */
@route("/hub/team/onboarding/v1")
interface OnboardingService {
  /** Marks the current user as onboarded. */
  @patch
  @route("/onboarded")
  markOnboarded(...MarkOnboardedRequest): MarkOnboardedResponse;
}
