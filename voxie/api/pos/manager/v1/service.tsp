import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.manager.v1;

/** ManagerService defines POS manager override endpoints. */
@route("/pos/manager/v1")
interface ManagerService {
  /** Verifies a manager PIN and returns the associated staff member. */
  @post
  @route("/verify")
  verifyPIN(...VerifyPINRequest): VerifyPINResponse;

  /** Sets or rotates the manager override PIN for a staff member. */
  @post
  @route("/set-pin")
  setPIN(...SetPINRequest): SetPINResponse;
}
