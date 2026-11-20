import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.setting.v1;

@route("/flow/setting/v1")
interface SettingService {
  @get
  @route("/settings")
  getUserSettings(...GetUserSettingsRequest): GetUserSettingsResponse;

  @put
  @route("/settings")
  updateUserSettings(...UpdateUserSettingsRequest): UpdateUserSettingsResponse;
}
