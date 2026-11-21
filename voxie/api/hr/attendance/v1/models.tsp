import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.attendance.v1;

/** AttendanceRecordPayload is a single clock-in/clock-out attendance record. */
model AttendanceRecordPayload {
  /** Unique identifier for the attendance record. */
  @encode(string)
  id: int64;

  /** Tenant owner this record belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee this attendance record is for. */
  @encode(string)
  employee_id: int64;

  /** Work location where attendance was recorded. */
  @encode(string)
  location_id: int64;

  /** Date of the attendance record (ISO-8601 format). */
  date: string;

  /** Unix timestamp when the employee clocked in. */
  @encode(string)
  clock_in_at: int64;

  /** Unix timestamp when the employee clocked out; absent if still clocked in. */
  @encode(string)
  clock_out_at?: int64;

  /** Total worked duration in minutes; absent if still clocked in. */
  duration_minutes?: int32;

  /** Method used to clock in (app | qr | gps | pin | facial). */
  clock_in_method: string;

  /** Latitude at clock-in; absent if location not captured. */
  clock_in_lat?: float64;

  /** Longitude at clock-in; absent if location not captured. */
  clock_in_lng?: float64;

  /** Whether the employee clocked in after the scheduled start. */
  is_late: boolean;

  /** Number of minutes late; zero if on time. */
  late_minutes: int32;

  /** Attendance status (present | absent | partial). */
  status: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model AttendanceClockInRequest {
  /** Method used to clock in (app | qr | gps | pin | facial). */
  @Onkt.required
  method: string;

  /** Work location where the employee is clocking in. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  location_id?: int64;

  /** Latitude at clock-in; absent if location not available. */
  lat?: float64;

  /** Longitude at clock-in; absent if location not available. */
  lng?: float64;

  /** QR token scanned at a terminal; absent if not using QR method. */
  qr_token?: string;

  /** PIN entered by the employee; absent if not using PIN method. Sensitive: do not log this field. */
  pin?: string;
}

model AttendanceClockInResponse {
  /** The created attendance record. */
  record: AttendanceRecordPayload;
}

model AttendanceClockOutRequest {
  /** ID of the attendance record to close. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  attendance_record_id?: int64;
}

model AttendanceClockOutResponse {
  /** The updated attendance record with clock-out data. */
  record: AttendanceRecordPayload;
}

model ListAttendanceRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by date; absent to include all dates. */
  @query("date")
  date?: string;
}

model ListAttendanceResponse {
  /** Attendance records matching the filters. */
  records: AttendanceRecordPayload[];

  /** Total number of matching attendance records across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}
