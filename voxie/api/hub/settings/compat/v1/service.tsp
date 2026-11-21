import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.settings.compat.v1;

/** Compatibility contract for the SMTP connection test used by Dash. */
@route("/hub/settings")
interface CompatibilityService {
  @post
  @route("/email/test")
  testEmail(...TestEmailRequest): TestEmailResponse;
}
