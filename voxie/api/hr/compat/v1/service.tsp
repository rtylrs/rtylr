import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.compat.v1;

/** Compatibility contract for the profile projection used by Flow and HR. */
@route("/hr")
interface CompatibilityService {
  @get
  @route("/employees/me")
  getMyEmployee(...EmptyRequest): JsonResponse;

  @get
  @route("/time-off")
  listTimeOff(...QueryRequest): JsonResponse;

  @post
  @route("/time-off")
  submitTimeOff(...BodyRequest): JsonResponse;

  @post
  @route("/time-off/approve")
  approveTimeOff(...BodyRequest): JsonResponse;

  @get
  @route("/shifts")
  getShifts(...QueryRequest): JsonResponse;

  @get
  @route("/shifts/availability")
  getAvailability(...QueryRequest): JsonResponse;

  @post
  @route("/shifts")
  upsertShift(...BodyRequest): JsonResponse;

  @post
  @route("/shifts/availability")
  setAvailability(...BodyRequest): JsonResponse;

  @put
  @route("/shifts")
  updateShift(...BodyRequest): JsonResponse;

  @get
  @route("/shifts/roster")
  getRoster(...QueryRequest): JsonResponse;

  @post
  @route("/shifts/roster/assignments")
  createShiftAssignment(...BodyRequest): JsonResponse;

  @post
  @route("/shifts/roster/publish")
  publishRoster(...BodyRequest): JsonResponse;

  @get
  @route("/timesheets")
  getTimesheets(...QueryRequest): JsonResponse;

  @get
  @route("/timesheets/current")
  getCurrentTimesheet(...EmptyRequest): JsonResponse;

  @get
  @route("/timesheets/{id}")
  getTimesheet(...PathRequest): JsonResponse;

  @post
  @route("/timesheets/{id}/submit")
  submitTimesheet(...PathBodyRequest): JsonResponse;

  @post
  @route("/timesheets/{id}/approve")
  approveTimesheet(...PathBodyRequest): JsonResponse;

  @post
  @route("/timesheets/{id}/reject")
  rejectTimesheet(...PathBodyRequest): JsonResponse;

  @get
  @route("/documents")
  getDocuments(...QueryRequest): JsonResponse;

  @get
  @route("/documents/expiring")
  getExpiringDocuments(...QueryRequest): JsonResponse;

  @post
  @route("/documents")
  createDocument(...BodyRequest): JsonResponse;

  @delete
  @route("/documents/{id}")
  deleteDocument(...PathRequest): JsonResponse;

  @get
  @route("/config")
  getConfig(...EmptyRequest): JsonResponse;

  @put
  @route("/config")
  updateConfig(...BodyRequest): JsonResponse;

  @get
  @route("/employees")
  getEmployees(...QueryRequest): JsonResponse;

  @get
  @route("/employees/{id}")
  getEmployee(...PathRequest): JsonResponse;

  @post
  @route("/employees")
  createEmployee(...BodyRequest): JsonResponse;

  @put
  @route("/employees/{id}")
  updateEmployee(...PathBodyRequest): JsonResponse;

  @post
  @route("/employees/invite")
  inviteEmployee(...BodyRequest): JsonResponse;

  @get
  @route("/attendance")
  getAttendance(...QueryRequest): JsonResponse;

  @put
  @route("/employees/me/pin")
  updateEmployeePin(...BodyRequest): JsonResponse;

  @get
  @route("/roles")
  getRoles(...EmptyRequest): JsonResponse;

  @get
  @route("/permissions")
  getPermissions(...EmptyRequest): JsonResponse;

  @post
  @route("/roles")
  createRole(...BodyRequest): JsonResponse;

  @put
  @route("/roles/{id}")
  updateRole(...PathBodyRequest): JsonResponse;

  @get
  @route("/performance/cycles/{id}")
  getReviewCycle(...PathRequest): JsonResponse;

  @post
  @route("/performance/cycles/{id}/open")
  openReviewCycle(...PathBodyRequest): JsonResponse;

  @post
  @route("/performance/goals")
  createGoal(...BodyRequest): JsonResponse;

  @put
  @route("/performance/goals/{id}")
  updateGoal(...PathBodyRequest): JsonResponse;

  @get
  @route("/performance/reviews")
  getPerformanceReviews(...QueryRequest): JsonResponse;

  @post
  @route("/performance/reviews/{id}/submit")
  submitPerformanceReview(...PathBodyRequest): JsonResponse;

  @post
  @route("/performance/reviews/{id}/share")
  sharePerformanceReview(...PathBodyRequest): JsonResponse;

  @get
  @route("/leave/calendar")
  getLeaveCalendar(...QueryRequest): JsonResponse;

  @post
  @route("/locations")
  createLocation(...BodyRequest): JsonResponse;

  @get
  @route("/setup/status")
  getSetupStatus(...EmptyRequest): JsonResponse;

  @post
  @route("/setup/steps/{id}/complete")
  completeSetupStep(...PathBodyRequest): JsonResponse;

  @get
  @route("/overtime/config")
  getOvertimeConfig(...EmptyRequest): JsonResponse;

  @put
  @route("/overtime/config")
  updateOvertimeConfig(...BodyRequest): JsonResponse;

  @get
  @route("/payroll/config")
  getPayrollConfig(...EmptyRequest): JsonResponse;

  @put
  @route("/payroll/config")
  updatePayrollConfig(...BodyRequest): JsonResponse;

  @get
  @route("/payroll/runs")
  getPayrollRuns(...QueryRequest): JsonResponse;

  @get
  @route("/payroll/runs/{id}")
  getPayrollRun(...PathRequest): JsonResponse;

  @get
  @route("/payroll/my-runs")
  getMyPayrollRuns(...QueryRequest): JsonResponse;

  @get
  @route("/payroll/my-runs/{id}")
  getMyPayrollRun(...PathRequest): JsonResponse;

  @post
  @route("/payroll/runs/trigger")
  triggerPayrollRun(...BodyRequest): JsonResponse;

  @post
  @route("/payroll/runs/{id}/approve")
  approvePayrollRun(...PathBodyRequest): JsonResponse;

  @get
  @route("/time-entries")
  getTimeEntries(...QueryRequest): JsonResponse;

  @post
  @route("/time-entries")
  createTimeEntry(...BodyRequest): JsonResponse;

  @put
  @route("/time-entries/{id}")
  updateTimeEntry(...PathBodyRequest): JsonResponse;

  @delete
  @route("/time-entries/{id}")
  deleteTimeEntry(...PathRequest): JsonResponse;

  @patch
  @route("/time-entries/mark-invoiced")
  markTimeEntriesInvoiced(...BodyRequest): JsonResponse;

  @delete
  @route("/employees/{id}")
  deleteEmployee(...PathIdRequest): JsonResponse;

  @delete
  @route("/shifts")
  deleteShift(...EmptyRequest): JsonResponse;

  @delete
  @route("/shifts/availability")
  deleteAvailability(...EmptyRequest): JsonResponse;

  @get
  @route("/company-config")
  getCompanyConfig(...EmptyRequest): JsonResponse;

  @get
  @route("/core/departments")
  listCoreDepartments(...EmptyRequest): JsonResponse;

  @get
  @route("/documents/{id}")
  getDocument(...PathIdRequest): JsonResponse;

  @get
  @route("/employees/invite/verify")
  verifyEmployeeInvite(...EmptyRequest): JsonResponse;

  @get
  @route("/locations")
  listLocations(...EmptyRequest): JsonResponse;

  @get
  @route("/onboarding/checklists/{id}")
  getEmployeeChecklist(...PathIdRequest): JsonResponse;

  @get
  @route("/performance/config")
  getPerformanceConfig(...EmptyRequest): JsonResponse;

  @post
  @route("/core/departments")
  createCoreDepartment(...BodyRequest): JsonResponse;

  @post
  @route("/employees/invite/accept")
  acceptEmployeeInvite(...BodyRequest): JsonResponse;

  @post
  @route("/employees/{id}/role")
  assignEmployeeRole(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/onboarding/checklists/items/{id}/complete")
  completeChecklistItem(...PathIdBodyRequest): JsonResponse;

  @post
  @route("/payroll/runs/{id}/export")
  exportPayrollRun(...PathIdBodyRequest): JsonResponse;

  @put
  @route("/company-config")
  updateCompanyConfig(...BodyRequest): JsonResponse;

  @put
  @route("/core/departments/{id}")
  updateCoreDepartment(...PathIdBodyRequest): JsonResponse;

  @put
  @route("/locations/{id}")
  updateLocation(...PathIdBodyRequest): JsonResponse;

  @put
  @route("/performance/config")
  updatePerformanceConfig(...BodyRequest): JsonResponse;

  @post
  @route("/reports/generate")
  @Onkt.http
  createReportsGenerate(...EmptyRequest): EmptyResponse;
}
