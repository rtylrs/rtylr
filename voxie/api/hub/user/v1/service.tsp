import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.user.v1;

/** UserService defines hub user profile endpoints. */
@route("/hub/user/v1")
interface UserService {
  /** Returns the authenticated user profile. */
  @get
  @route("/users/me")
  getMe(...GetMeRequest): GetMeResponse;

  /** Updates the authenticated user profile. */
  @patch
  @route("/users/me")
  updateProfile(...UpdateProfileRequest): UpdateProfileResponse;

  /** Sets or resets the authenticated user's avatar. */
  @put
  @route("/users/me/avatar")
  updateAvatar(...UpdateAvatarRequest): UpdateAvatarResponse;

  /** Lists the chosen avatars of the caller's business colleagues. */
  @get
  @route("/users/avatars")
  listBusinessAvatars(
    ...ListBusinessAvatarsRequest,
  ): ListBusinessAvatarsResponse;

  /** Changes the authenticated user password. */
  @post
  @route("/users/password")
  changePassword(...ChangePasswordRequest): ChangePasswordResponse;

  /** Issues a fresh JWT scoped to the selected business. */
  @post
  @route("/users/switch-business")
  switchBusiness(...SwitchBusinessRequest): SwitchBusinessResponse;
}
