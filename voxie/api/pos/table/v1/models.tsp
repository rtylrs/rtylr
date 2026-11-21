import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.table.v1;

/** Table is a POS floor table. */
model Table {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  @encode(string)
  location_id?: int64;

  label: string;
  seats: int32;

  /** status is the current lifecycle state. */
  status: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model ListTablesRequest {}

model ListTablesResponse {
  tables: Table[];
}

model CreateTableRequest {
  @encode(string)
  location_id?: int64;

  @Onkt.required
  label: string;

  @Onkt.required
  @minValueExclusive(0)
  seats?: int32;
}

model CreateTableResponse {
  @encode(string)
  table_id: int64;

  message: string;
}

model UpdateTableBody {
  @encode(string)
  location_id?: int64;

  @Onkt.required
  label: string;

  @Onkt.required
  @minValueExclusive(0)
  seats?: int32;

  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateTableRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable table fields bound from the HTTP payload. */
  @body
  body: UpdateTableBody;
}

model UpdateTableResponse {
  message: string;
}

model UpdateTableStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateTableStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new table status bound from the HTTP payload. */
  body: UpdateTableStatusBody;
}

model UpdateTableStatusResponse {
  message: string;
}

model DeleteTableRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteTableResponse {
  message: string;
}
