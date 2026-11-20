import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.compat.v1;

model EmptyRequest {}

/** Empty response for \@http routes whose handler writes its own response. */
model EmptyResponse {}

model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}

model QueryRequest {
  @query("slug")
  slug?: string;

  @query("career_slug")
  career_slug?: string;

  @query("career_id")
  career_id?: int32;

  @query("submission_id")
  submission_id?: int32;

  @query("status")
  status?: string;

  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;
}

model BodyRequest {
  @body
  body: unknown;
}

model IdQueryRequest {
  @query("id")
  id: string;
}

model PathIdRequest {
  id: string;
}

model PathSourceIdRequest {
  source: string;
  id: string;
}
