import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.bill.v1;

/** Bill is an accounts-payable bill owed to a supplier. */
model Bill {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /**
   * supplier_id is the ERP supplier this bill is owed to; absent on bills
   * entered before suppliers were linked, and on one-off bills for a party
   * with no supplier record. Callers must not invent a value: a bill with no
   * known supplier records none.
   */
  @encode(string)
  supplier_id?: int64;

  /**
   * supplier_name is the supplier's name as it stood when the bill was
   * entered, so a bill list reads correctly without a second request and a
   * renamed supplier does not rewrite history. Required: every bill records
   * who it is owed to, by reference or by snapshot.
   */
  @Onkt.required
  supplier_name: string;

  /**
   * purchase_order_id is the ERP purchase order this bill settles; absent for
   * bills that did not originate from a purchase order.
   */
  @encode(string)
  purchase_order_id?: int64;

  /** amount is the amount value for this contract. */
  amount: float64;

  /** currency is the currency value for this contract. */
  currency: string;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** status is the current lifecycle state. */
  status: string;

  /** description is the description value for this contract. */
  description: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateBillRequest {
  /**
   * supplier_id is the ERP supplier this bill is owed to. Optional: supply it
   * from ERP's supplier list when the party has a record, and omit it for a
   * one-off bill. Callers must never send a derived or placeholder value.
   */
  @encode(string)
  supplier_id?: int64;

  /**
   * supplier_name is the vendor name as typed or picked at entry time. It is
   * the display snapshot for the bill and must always be present even when
   * supplier_id is set, so history survives a later rename.
   */
  @Onkt.required
  supplier_name: string;

  /**
   * purchase_order_id links this bill to the ERP purchase order it settles;
   * absent when the bill did not originate from a purchase order. The server
   * must verify the referenced order belongs to the caller's business.
   */
  @encode(string)
  purchase_order_id?: int64;

  /** amount is the amount value for this contract. */
  @minValue(0)
  amount: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** due_date is the due date value for this contract. */
  @Onkt.required
  due_date: string;

  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;
}

model CreateBillResponse {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model DeleteBillRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteBillResponse {
  message: string;
}

model GetBillRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetBillResponse {
  bill: Bill;
}

model ListBillsRequest {
  /**
   * supplier_id filters to one ERP supplier's bills; absent to include all.
   * Absent supplier links are excluded when this filter is supplied.
   */
  @query("supplier_id")
  @encode(string)
  supplier_id?: int64;

  /**
   * purchase_order_id filters to bills settling one purchase order; absent
   * to include all.
   */
  @query("purchase_order_id")
  @encode(string)
  purchase_order_id?: int64;

  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListBillsResponse {
  /** bills is the bills value for this contract. */
  bills: Bill[];

  page: int64;
  page_size: int64;
  has_more: boolean;
}

model UpdateBillBody {
  /**
   * supplier_id is the ERP supplier this bill is owed to; absent keeps the
   * stored value, and a one-off bill may clear it by sending nothing here.
   */
  @encode(string)
  supplier_id?: int64;

  /** supplier_name replaces the vendor snapshot; absent means unchanged. */
  supplier_name: string;

  /**
   * purchase_order_id re-links this bill to an ERP purchase order; absent
   * keeps the stored link.
   */
  @encode(string)
  purchase_order_id?: int64;

  /** amount is the amount value for this contract. */
  @minValue(0)
  amount: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** due_date is the due date value for this contract. */
  @Onkt.required
  due_date: string;

  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;
}

model UpdateBillRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable bill fields bound from the HTTP payload. */
  @body
  body: UpdateBillBody;
}

model UpdateBillResponse {
  /** message is a human-readable result message. */
  message: string;
}
