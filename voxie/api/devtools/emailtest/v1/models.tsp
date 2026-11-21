import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace devtools.emailtest.v1;

model ListTemplatesRequest {}

model ListTemplatesResponse {
  count: int32;
  templates: string[];
}

/**
 * data is optional: missing template fields render empty, so a bare
 * {"template","to"} is enough to test delivery.
 */
model SendTestEmailRequest {
  template: string;
  to: string;
  subject: string;
  data: unknown;
}

model SendTestEmailResponse {
  queued: boolean;
  template: string;
  to: string;
}
