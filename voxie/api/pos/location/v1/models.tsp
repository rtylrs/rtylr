import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.location.v1;

/** Location is a physical POS location for a business. */
model Location {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  name: string;
  address: string;
  phone: string;
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateLocationRequest {
  @Onkt.required
  name: string;

  address: string;
  phone: string;

  /** Defaults to true when absent. */
  is_active?: boolean;
}

model CreateLocationResponse {
  @encode(string)
  location_id: int64;

  message: string;
}

model ListLocationsRequest {}

model ListLocationsResponse {
  locations: Location[];
}

model UpdateLocationBody {
  @Onkt.required
  name: string;

  address: string;
  phone: string;

  /** Unchanged when absent. */
  is_active?: boolean;
}

model UpdateLocationRequest {
  /** id is the unique identifier for the location. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable location fields bound from the HTTP payload. */
  @body
  body: UpdateLocationBody;
}

model UpdateLocationResponse {
  message: string;
}
