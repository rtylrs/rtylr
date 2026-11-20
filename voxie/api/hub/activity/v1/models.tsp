import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.activity.v1;

model ActivityEntry {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  app: string;
  entity_type: string;

  @encode(string)
  entity_id: int64;

  action: string;
  description: string;

  @encode(string)
  user_id: int64;

  user_name: string;

  /** When the record was created. */
  created_at: string;
}

model ListActivityRequest {
  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** offset is the number of records to skip, for paging through the feed. */
  @query("offset")
  offset?: int32;
}

model ListActivityResponse {
  activities: ActivityEntry[];

  /** total is the total number of matching activity entries, for pagination math. */
  total: int32;
}
