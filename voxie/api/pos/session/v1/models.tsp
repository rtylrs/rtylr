import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.session.v1;

/** Session is a POS register session. */
model Session {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  @encode(string)
  location_id?: int64;

  @encode(string)
  staff_id?: int64;

  opening_float: float64;
  closing_float: float64;
  total_sales: float64;

  /** status is the current lifecycle state. */
  status: string;

  opened_at: string;
  closed_at?: string;

  /** created_at is when the record was created. */
  created_at: string;
}

/** CurrentSession is the active open register session for the requesting staff member. */
model CurrentSession {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the session. */
  @encode(string)
  owner_id: int64;

  /** location_id is the active location id. */
  @encode(string)
  location_id: int64;

  /** staff_id is the staff member who opened the session. */
  @encode(string)
  staff_id: int64;

  /** opening_cash is the cash drawer opening float. */
  opening_cash: float64;

  /** closing_cash is the cash drawer closing float. */
  closing_cash: float64;

  /** total_sales is the current session sales total. */
  total_sales: float64;

  /** status is the current lifecycle state. */
  status: string;

  /** opened_at is when the session was opened. */
  opened_at: string;

  /** closed_at is when the session was closed. */
  closed_at?: string;
}

model GetSessionRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetSessionResponse {
  session: Session;
}

model GetCurrentSessionRequest {}

model GetCurrentSessionResponse {
  /** session is the current open register session; null when none is open. */
  session: CurrentSession;
}

model ListSessionsRequest {}

model ListSessionsResponse {
  sessions: Session[];
}

model OpenSessionRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id?: int64;

  @Onkt.required
  @minValue(0)
  opening_float?: float64;
}

model OpenSessionResponse {
  @encode(string)
  session_id: int64;

  message: string;
}

model CloseSessionRequest {
  @Onkt.required
  @minValue(0)
  closing_float?: float64;

  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model CloseSessionResponse {
  message: string;
}
