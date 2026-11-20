import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hr.staff.v1;

namespace hr.employee.v1;

/** EmployeePayload is an HR employee record. */
model EmployeePayload {
  /** Unique identifier for the employee record. */
  @encode(string)
  id: int64;

  /** Tenant owner this employee belongs to. */
  @encode(string)
  owner_id: int64;

  /** Linked user account ID. */
  @encode(string)
  user_id: int64;

  /** Department the employee is assigned to. */
  @encode(string)
  department_id: int64;

  /** Work location the employee reports to. */
  @encode(string)
  location_id: int64;

  /** Direct manager; absent if top-level. */
  @encode(string)
  manager_id?: int64;

  /** Employment arrangement (full_time | part_time | freelancer). */
  contract_type: string;

  /** Current employment status (active | probation | inactive | terminated). */
  status: string;

  /** Date the employee was hired (ISO-8601 format). */
  hire_date: string;

  /** End of probation period; absent if no probation (ISO-8601 format). */
  probation_end_date?: string;

  /** Date employment was terminated; absent if still employed (ISO-8601 format). */
  termination_date?: string;

  /** Pay rate per hour for hourly employees; absent if salaried only. */
  hourly_rate?: float64;

  /** Monthly gross salary; absent if hourly only. */
  monthly_salary?: float64;

  /** ISO-4217 currency code for compensation (e.g. "USD"). */
  currency: string;

  /** Assigned work schedule; absent if using department default. */
  @encode(string)
  work_schedule_id?: int64;

  /** How the employee account was created (hr_created | invite_link). */
  account_creation_mode: string;

  /** Employee's job title or position name. */
  job_title: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model CreateEmployeeRequest {
  /** user_id links this employee record to an existing login account; zero leaves it unlinked. */
  @encode(string)
  user_id: int64;

  /** department_id is the department assignment; zero leaves it unset. */
  @encode(string)
  department_id: int64;

  /** location_id is the work location assignment; zero leaves it unset. */
  @encode(string)
  location_id: int64;

  /** work_schedule_id is the work schedule assignment; zero leaves it unset. */
  @encode(string)
  work_schedule_id: int64;

  /** first_name is the employee's given name. */
  @Onkt.required
  first_name: string;

  /** last_name is the employee's family name. */
  @Onkt.required
  last_name: string;

  /** email is the employee email address. */
  @Onkt.required
  @format("email")
  email: string;

  /** phone is the employee phone number. */
  phone: string;

  /** employment_type is the employment type label. */
  @Onkt.required
  employment_type: string;

  /** salary is the monthly or salaried compensation amount. */
  @minValue(0)
  salary: float64;

  /** hourly_rate is the employee hourly pay rate. */
  @minValue(0)
  hourly_rate: float64;

  /** overtime_rate is the employee overtime pay rate. */
  @minValue(0)
  overtime_rate: float64;

  /** start_date is the start date in YYYY-MM-DD format. */
  @Onkt.required
  start_date: string;

  /** address is the employee address. */
  address: string;
}

model CreateEmployeeResponse {
  /** The newly created employee record. */
  employee: Staff;
}

model InviteEmployeeRequest {
  /** ID of the employee to invite. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** Channel to deliver the invite (email | sms). */
  @Onkt.required
  send_via: string;
}

model InviteEmployeeResponse {
  /** Sensitive: do not log this field. One-time invite token for the employee. */
  invite_token: string;

  /** Unix timestamp when the invite expires. */
  @encode(string)
  expires_at: int64;
}

model AcceptInviteRequest {
  /** Sensitive: do not log this field. The invite token received by the employee. */
  @Onkt.required
  token: string;

  /** Sensitive: do not log this field. Password chosen by the employee. */
  @Onkt.required
  password: string;
}

model AcceptInviteResponse {
  /** Whether the invite was accepted successfully. */
  success: boolean;
}

model ListEmployeesRequest {
  /** Filter by department; absent to include all departments. */
  @query("department_id")
  @encode(string)
  department_id?: int64;

  /** Filter by location; absent to include all locations. */
  @query("location_id")
  @encode(string)
  location_id?: int64;

  /** Filter by employment status; absent to include all statuses. */
  @query("status")
  status?: string;

  /** Maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** Number of records to skip for pagination. */
  @query("offset")
  offset: int32;
}

model ListEmployeesResponse {
  /** Page of employee records matching the filters. */
  employees: EmployeePayload[];

  /** Total number of employees matching the filters. */
  total: int64;

  /** Limit that was applied. */
  limit: int32;

  /** Offset that was applied. */
  offset: int32;
}

model UpdateEmployeeBody {
  /** user_id links this employee record to an existing login account; zero leaves it unlinked. */
  @encode(string)
  user_id: int64;

  /** department_id is the department assignment; zero leaves it unset. */
  @encode(string)
  department_id: int64;

  /** location_id is the work location assignment; zero leaves it unset. */
  @encode(string)
  location_id: int64;

  /** work_schedule_id is the work schedule assignment; zero leaves it unset. */
  @encode(string)
  work_schedule_id: int64;

  /** first_name is the employee's given name. */
  @Onkt.required
  first_name: string;

  /** last_name is the employee's family name. */
  @Onkt.required
  last_name: string;

  /** email is the employee email address. */
  @Onkt.required
  email: string;

  /** phone is the employee phone number. */
  phone: string;

  /** employment_type is the employment type label. */
  @Onkt.required
  employment_type: string;

  /** status is the employment status. */
  @Onkt.required
  status: string;

  /** salary is the monthly or salaried compensation amount. */
  @minValue(0)
  salary: float64;

  /** hourly_rate is the employee hourly pay rate. */
  @minValue(0)
  hourly_rate: float64;

  /** overtime_rate is the employee overtime pay rate. */
  @minValue(0)
  overtime_rate: float64;

  /** start_date is the start date in YYYY-MM-DD format. */
  @Onkt.required
  start_date: string;

  /** end_date is the end date in YYYY-MM-DD format. */
  @Onkt.required
  end_date: string;

  /** address is the employee address. */
  address: string;

  /** emergency_contact_name is the emergency contact name. */
  @Onkt.required
  emergency_contact_name: string;

  /** emergency_contact_phone is the emergency contact phone. */
  @Onkt.required
  emergency_contact_phone: string;

  /** notes are internal employee notes. */
  notes: string;
}

model UpdateEmployeeRequest {
  /** id is the employee ID from the route. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable employee fields bound from the HTTP payload. */
  @body
  body: UpdateEmployeeBody;
}

model UpdateEmployeeResponse {
  /** The updated employee record. */
  employee: Staff;
}
