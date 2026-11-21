import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.settings.v1;

/** Settings holds CRM-wide messaging defaults for a business. */
model Settings {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  default_consent: boolean;
  from_email: string;
  from_name: string;
  sms_sender_id: string;
  timezone: string;
  created_at: string;
  updated_at: string;
}

model GetSettingsRequest {}

model GetSettingsResponse {
  settings: Settings;
}

model UpdateSettingsRequest {
  default_consent: boolean;
  from_email: string;
  from_name: string;
  sms_sender_id: string;
  timezone: string;
}

model UpdateSettingsResponse {
  message: string;
}
