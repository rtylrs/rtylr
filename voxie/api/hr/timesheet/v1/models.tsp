import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.timesheet.v1;

/** TimesheetEntry is a weekly hours summary for a staff member. */
model TimesheetEntry {
  staff_id: int64;
  staff_name: string;
  week_start: string;
  total_hours: float64;
  regular_hours: float64;
  overtime_hours: float64;
}

model ListTimesheetsRequest {
  @query("week_start")
  week_start: string;
}

model ListTimesheetsResponse {
  timesheets: TimesheetEntry[];
}
