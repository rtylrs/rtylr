import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.settings.v1;

/** SettingsService defines tenant-scoped ERP business settings endpoints. */
@route("/erp/settings/v1")
interface SettingsService {
  @get
  @route("/settings")
  getErpSettings(...GetErpSettingsRequest): GetErpSettingsResponse;

  @put
  @route("/settings")
  updateErpSettings(...UpdateErpSettingsRequest): UpdateErpSettingsResponse;
}
