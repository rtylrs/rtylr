import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.settings.v1;

/** SettingsService defines POS business settings endpoints. */
@route("/pos/settings/v1")
interface SettingsService {
  /** Returns the active business POS settings. */
  @get
  @route("/settings")
  getPOSSettings(...GetPOSSettingsRequest): GetPOSSettingsResponse;

  /** Updates the active business POS settings. */
  @put
  @route("/settings")
  updatePOSSettings(...UpdatePOSSettingsRequest): UpdatePOSSettingsResponse;
}
