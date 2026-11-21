import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.reminder.v1;

model Reminder {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  appointment_id: int64;

  type: string;
  scheduled_at: string;
  sent_at: string;
  created_at: string;
}

model CreateReminderRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  appointment_id?: int64;

  @Onkt.required
  type: string;

  @Onkt.required
  scheduled_at: string;
}

model CreateReminderResponse {
  @encode(string)
  id: int64;

  message: string;
}

model ListRemindersRequest {}

model ListRemindersResponse {
  reminders: Reminder[];
}
