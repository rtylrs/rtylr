import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.calendar.v1;

@route("/booking/calendar/v1")
interface CalendarService {
  @get
  @route("/calendar")
  getCalendar(...GetCalendarRequest): GetCalendarResponse;
}
