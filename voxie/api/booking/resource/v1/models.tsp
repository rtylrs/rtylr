import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.resource.v1;

model Resource {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  name: string;
  type: string;
  capacity: int32;
  is_active: boolean;
  created_at: string;
}

model CreateResourceRequest {
  @Onkt.required
  name: string;

  type: string;
  capacity: int32;
  is_active: boolean;
}

model CreateResourceResponse {
  @encode(string)
  id: int64;

  message: string;
}

model DeleteResourceRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteResourceResponse {
  message: string;
}

model ListResourcesRequest {}

model ListResourcesResponse {
  resources: Resource[];
}

model UpdateResourceRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  name: string;
  type: string;
  capacity: int32;
  is_active: boolean;
}

model UpdateResourceResponse {
  message: string;
}
