import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace store.compat.v1;

model EmptyRequest {}

/** The route's JSON body, passed through unchanged. */
model BodyRequest {
  @body
  body: unknown;
}

/** The handler's JSON response, sent unchanged. */
model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}
