import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.manager.v1;

model VerifyPINRequest {
  @Onkt.required
  pin: string;
}

model VerifyPINResponse {
  valid: boolean;

  @encode(string)
  staff_id: int64;

  message: string;
}

model SetPINRequest {
  /** staff_id is the manager the override is attributed to. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id: int64;

  /** pin is the new manager override PIN, 6-12 digits. Sensitive: do not log this field. */
  @Onkt.required
  pin: string;

  /**
   * current_password re-authenticates the business owner, so a stolen or
   * left-open session cannot mint a manager PIN. Sensitive: do not log this field.
   */
  @Onkt.required
  current_password: string;
}

model SetPINResponse {
  success: boolean;
}
