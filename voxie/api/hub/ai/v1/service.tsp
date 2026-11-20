import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.ai.v1;

@route("/hub/ai/v1")
interface AiService {
  @get
  @route("/usage")
  getAiUsage(...GetAiUsageRequest): GetAiUsageResponse;

  @get
  @route("/quota-settings")
  getAiQuotaSettings(...GetAiQuotaSettingsRequest): GetAiQuotaSettingsResponse;

  @put
  @route("/quota-settings")
  updateAiQuotaSettings(
    ...UpdateAiQuotaSettingsRequest,
  ): UpdateAiQuotaSettingsResponse;
}
