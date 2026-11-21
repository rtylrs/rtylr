import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.calendar.v1;

model CalendarAppointment {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  service_id: int64;

  customer_name: string;
  start_time: string;
  end_time: string;
  status: string;
}

model CalendarDay {
  date: string;
  appointments: CalendarAppointment[];
}

model GetCalendarRequest {
  @Onkt.required
  @query("from")
  from_date: string;

  @Onkt.required
  @query("to")
  to_date: string;
}

model GetCalendarResponse {
  days: CalendarDay[];
}
