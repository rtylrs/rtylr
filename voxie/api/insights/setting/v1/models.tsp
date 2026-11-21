import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.setting.v1;

model InsightsSettings {
  /**
   * default_period is the default period value for this contract.
   * this_month, this_quarter, this_year
   */
  default_period: string;

  /** currency is the currency value for this contract. */
  currency: string;

  fiscal_year_start: string;
  timezone: string;

  /** compare_previous is the compare previous value for this contract. */
  compare_previous: boolean;

  /** dashboard_widgets is the dashboard widgets value for this contract. */
  dashboard_widgets: string[];

  /** favorite_reports is the favorite reports value for this contract. */
  favorite_reports: string[];
}

model GetInsightsSettingsRequest {}

model GetInsightsSettingsResponse {
  /** settings is the settings value for this contract. */
  settings: InsightsSettings;
}

model UpdateInsightsSettingsRequest {
  /** settings is the settings value for this contract. */
  @Onkt.required
  settings: InsightsSettings;
}

model UpdateInsightsSettingsResponse {
  /** settings is the settings value for this contract. */
  settings: InsightsSettings;
}
