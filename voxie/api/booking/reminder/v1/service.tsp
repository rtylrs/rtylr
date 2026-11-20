import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.reminder.v1;

@route("/booking/reminder/v1")
interface ReminderService {
  @get
  @route("/reminders")
  listReminders(...ListRemindersRequest): ListRemindersResponse;

  @post
  @route("/reminders")
  createReminder(...CreateReminderRequest): CreateReminderResponse;
}
