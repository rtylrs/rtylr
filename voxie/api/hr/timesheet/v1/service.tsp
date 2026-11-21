import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.timesheet.v1;

@route("/hr/timesheet/v1")
interface TimesheetService {
  @get
  @route("/timesheets")
  listTimesheets(...ListTimesheetsRequest): ListTimesheetsResponse;
}
