import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace organization.compat.v1;

model EmptyRequest {}

/** Empty response for \@http routes whose handler writes its own response. */
model EmptyResponse {}

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

model PathIdRequest {
  id: string;
}

model PathTableIdRequest {
  table: string;
  id: string;
}

model PathIdBodyRequest {
  id: string;

  @body
  body: unknown;
}

model PathTableBodyRequest {
  table: string;

  @body
  body: unknown;
}

model PathTableIdBodyRequest {
  table: string;
  id: string;

  @body
  body: unknown;
}

model PathTableRequest {
  table: string;
}
