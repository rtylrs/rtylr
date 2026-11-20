import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.setting.v1;

/** SettingService defines insights settings endpoints. */
@route("/insights/setting/v1")
interface SettingService {
  /** Returns insights configuration. */
  @get
  @route("/settings")
  getInsightsSettings(
    ...GetInsightsSettingsRequest,
  ): GetInsightsSettingsResponse;

  /** Replaces insights configuration. */
  @put
  @route("/settings")
  updateInsightsSettings(
    ...UpdateInsightsSettingsRequest,
  ): UpdateInsightsSettingsResponse;
}
