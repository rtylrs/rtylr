import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.floor.v1;

/** Floor is a physical seating area within a branch. */
model Floor {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  name: string;
  sort_order: int32;
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;
}

model CreateFloorRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  @minValueExclusive(0)
  sort_order?: int32;
}

model CreateFloorResponse {
  @encode(string)
  floor_id: int64;

  message: string;
}

model UpdateFloorBody {
  @Onkt.required
  name: string;

  @Onkt.required
  @minValueExclusive(0)
  sort_order?: int32;
}

model UpdateFloorRequest {
  /** id is the unique identifier for the floor. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable floor fields bound from the HTTP payload. */
  @body
  body: UpdateFloorBody;
}

model UpdateFloorResponse {
  message: string;
}

model DeleteFloorRequest {
  /** id is the unique identifier for the floor. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteFloorResponse {
  message: string;
}

model ListFloorsRequest {
  @query("branch_id")
  @encode(string)
  branch_id: int64;
}

model ListFloorsResponse {
  floors: Floor[];
}
