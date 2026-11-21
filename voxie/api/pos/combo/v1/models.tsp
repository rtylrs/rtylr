import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.combo.v1;

/** Combo is a bundled set of menu items sold as a single offering. */
model Combo {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  name: string;
  description: string;
  price: float64;
  items: string;
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateComboRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  description: string;

  @minValue(0)
  price: float64;

  @Onkt.required
  items: string;

  is_active: boolean;
}

model CreateComboResponse {
  @encode(string)
  combo_id: int64;

  message: string;
}

model UpdateComboBody {
  @Onkt.required
  name: string;

  @Onkt.required
  description: string;

  @minValue(0)
  price: float64;

  @Onkt.required
  items: string;

  is_active: boolean;
}

model UpdateComboRequest {
  /** id is the unique identifier for the combo. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable combo fields bound from the HTTP payload. */
  @body
  body: UpdateComboBody;
}

model UpdateComboResponse {
  message: string;
}

model DeleteComboRequest {
  /** id is the unique identifier for the combo. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteComboResponse {
  message: string;
}

model ListCombosRequest {}

model ListCombosResponse {
  combos: Combo[];
}
