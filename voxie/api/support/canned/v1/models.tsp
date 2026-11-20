import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.canned.v1;

model CannedResponse {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  name: string;
  body: string;
  category: string;
  shortcut: string;
  created_at: string;
}

model ListCannedResponsesRequest {}

model ListCannedResponsesResponse {
  canned_responses: CannedResponse[];
}

model CreateCannedResponseRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  body: string;

  category: string;
  shortcut: string;
}

model CreateCannedResponseResponse {
  @encode(string)
  canned_response_id: int64;

  message: string;
}

model UpdateCannedResponseRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  name: string;
  body: string;
  category: string;
  shortcut: string;
}

model UpdateCannedResponseResponse {
  message: string;
}

model DeleteCannedResponseRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteCannedResponseResponse {
  message: string;
}
