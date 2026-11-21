import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.careers.compat.v1;

model EmptyRequest {}

/** Empty response for \@http routes whose handler writes its own response. */
model EmptyResponse {}

/** The route's JSON body, passed through unchanged. */
model BodyRequest {
  body: unknown;
}

/** The handler's JSON response, sent unchanged. */
model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}
