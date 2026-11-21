import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.user.v1;

/** UserService defines POS roster operator endpoints. */
@route("/pos/user/v1")
interface UserService {
  /** Returns POS roster operators for a location. */
  @get
  @route("/users")
  listPOSUsers(...ListPOSUsersRequest): ListPOSUsersResponse;

  /** Creates a new POS roster operator. */
  @post
  @route("/users")
  createPOSUser(...CreatePOSUserRequest): CreatePOSUserResponse;

  /** Updates an existing POS roster operator. */
  @patch
  @route("/users/update")
  updatePOSUser(...UpdatePOSUserRequest): UpdatePOSUserResponse;

  /** Deactivates a POS roster operator. */
  @patch
  @route("/users/deactivate")
  deactivatePOSUser(...DeactivatePOSUserRequest): DeactivatePOSUserResponse;

  /** Verifies an operator PIN for terminal sign-in. */
  @post
  @route("/users/verify-pin")
  verifyPOSUserPIN(...VerifyPOSUserPINRequest): VerifyPOSUserPINResponse;

  /** Sets or resets an operator PIN. */
  @post
  @route("/users/set-pin")
  setPIN(...SetPINRequest): SetPINResponse;
}
