import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.attendance.v1;

/** AttendanceService defines HR attendance endpoints. */
@route("/hr/attendance/v1")
interface AttendanceService {
  /** ListAttendance returns attendance records with optional filters. */
  @get
  @route("/attendance")
  listAttendance(...ListAttendanceRequest): ListAttendanceResponse;

  /** AttendanceClockIn records an attendance clock-in. */
  @post
  @route("/attendance/clock-in")
  attendanceClockIn(...AttendanceClockInRequest): AttendanceClockInResponse;

  /** AttendanceClockOut records an attendance clock-out. */
  @post
  @route("/attendance/clock-out")
  attendanceClockOut(...AttendanceClockOutRequest): AttendanceClockOutResponse;
}
