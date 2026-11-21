import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.modifier.v1;

/** Modifier is a customization option that can be attached to a menu item. */
model Modifier {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  name: string;
  price_delta: float64;
  group_name: string;
  is_required: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateModifierRequest {
  @Onkt.required
  name: string;

  @minValue(0)
  price_delta: float64;

  @Onkt.required
  group_name: string;

  is_required: boolean;
}

model CreateModifierResponse {
  @encode(string)
  modifier_id: int64;

  message: string;
}

model DeleteModifierRequest {
  /** id is the unique identifier for the modifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteModifierResponse {
  message: string;
}

model ListModifiersRequest {}

model ListModifiersResponse {
  modifiers: Modifier[];
}

model UpdateModifierBody {
  @Onkt.required
  name: string;

  @minValue(0)
  price_delta: float64;

  @Onkt.required
  group_name: string;

  is_required: boolean;
}

model UpdateModifierRequest {
  /** id is the unique identifier for the modifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable modifier fields bound from the HTTP payload. */
  @body
  body: UpdateModifierBody;
}

model UpdateModifierResponse {
  message: string;
}
