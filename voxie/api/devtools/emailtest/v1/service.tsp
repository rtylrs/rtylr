import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace devtools.emailtest.v1;

/**
 * Non-production endpoints for testing email templates end to end. They
 * answer 404 unless the run declares a development environment.
 */
@route("/")
interface EmailTestService {
  @get
  @route("/test/email/templates")
  listTemplates(...ListTemplatesRequest): ListTemplatesResponse;

  @post
  @route("/test/email")
  sendTestEmail(...SendTestEmailRequest): SendTestEmailResponse;
}
