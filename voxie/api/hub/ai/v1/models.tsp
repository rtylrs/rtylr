import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.ai.v1;

model AiQuotaStatus {
  plan_slug: string;
  mode: string;

  @encode(string)
  allowance_tokens: int64;

  @encode(string)
  used_tokens: int64;

  @encode(string)
  remaining_tokens: int64;

  @encode(string)
  overage_tokens: int64;

  allow_overage: boolean;
  blocked: boolean;
  period_end: string;
}

model AiQuotaSettings {
  mode: string;
}

model GetAiUsageRequest {}

model GetAiUsageResponse {
  usage: AiQuotaStatus;
}

model GetAiQuotaSettingsRequest {}

model GetAiQuotaSettingsResponse {
  settings: AiQuotaSettings;
}

model UpdateAiQuotaSettingsRequest {
  @Onkt.required
  mode: string;
}

model UpdateAiQuotaSettingsResponse {
  settings: AiQuotaSettings;
}
