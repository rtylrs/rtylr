import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.settings.v1;

@route("/support/settings/v1")
interface SettingsService {
  @get
  @route("/settings")
  getSettings(...GetSettingsRequest): GetSettingsResponse;

  @put
  @route("/settings")
  updateSettings(...UpdateSettingsRequest): UpdateSettingsResponse;
}
