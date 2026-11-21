import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.presence.v1;

model PresenceEntry {
  employee_id: string;
  user_id: string;
  first_name?: string;
  last_name?: string;
  status: string;
}

model GetPresenceRequest {}

model GetPresenceResponse {
  presence: PresenceEntry[];
  date: string;
}
