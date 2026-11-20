import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.settings.compat.v1;

model TestEmailBody {
  host: string;
  port: int32;
  username: string;
}

model TestEmailRequest {
  @body
  body: TestEmailBody;
}

model TestEmailResponse {
  @Onkt.unwrap
  data: unknown;
}
