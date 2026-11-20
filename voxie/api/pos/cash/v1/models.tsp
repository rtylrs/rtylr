import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.cash.v1;

/** CashMovement is a recorded cash drawer movement. */
model CashMovement {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  @encode(string)
  terminal_id?: int64;

  @encode(string)
  session_id: int64;

  @encode(string)
  employee_id: int64;

  type: string;
  amount: float64;
  reason?: string;

  @encode(string)
  approved_by?: int64;

  /** created_at is when the record was created. */
  created_at: string;
}

model CashAdjustRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  terminal_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  session_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  amount?: float64;

  @Onkt.required
  reason: string;
}

model CashAdjustResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model CashDropRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  terminal_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  session_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  amount?: float64;

  @Onkt.required
  reason: string;
}

model CashDropResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model CashPickupRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  terminal_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  session_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  amount?: float64;

  @Onkt.required
  reason: string;
}

model CashPickupResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model ListCashMovementsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("session_id")
  @encode(string)
  session_id?: int64;
}

model ListCashMovementsResponse {
  movements: CashMovement[];
}
