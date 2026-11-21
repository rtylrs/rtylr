import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.settings.v1;

model Settings {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  auto_assign: boolean;
  sla_hours: int32;
  business_hours: string;
  notification_email: string;
  created_at: string;
  updated_at: string;
}

model GetSettingsRequest {}

model GetSettingsResponse {
  settings: Settings;
}

model UpdateSettingsRequest {
  auto_assign: boolean;

  @Onkt.required
  @minValueExclusive(0)
  @maxValue(1000)
  sla_hours?: int32;

  business_hours: string;
  notification_email: string;
}

model UpdateSettingsResponse {
  message: string;
}
