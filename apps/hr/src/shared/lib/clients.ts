import { createOneKitClient, fromOneKitJson, oneKitResult, toOneKitJson } from '@voxire/api';
import { LicenseServiceClient } from '@voxire/voxie/license/v1/client';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { CompatibilityServiceClient } from '@voxire/voxie/hr/compat/v1/client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { ShiftServiceClient } from '@voxire/voxie/hr/shift/v1/client';

type HrRequestConfig = {
  params?: unknown;
  data?: unknown;
};

const hrCompatibilityClient = createOneKitClient(CompatibilityServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');
const shiftClient = createOneKitClient(ShiftServiceClient, 'hr');

type HrOperation = (request: unknown) => Promise<unknown>;

function invoke(method: string, request: unknown): Promise<unknown> {
  const operation = (hrCompatibilityClient as unknown as Record<string, HrOperation>)[method];
  if (!operation) throw new Error(`Unknown HR compatibility operation: ${method}`);
  return operation.call(hrCompatibilityClient, request);
}

function normalizeParams(value: unknown): Record<string, unknown> {
  return fromOneKitJson<Record<string, unknown>>(value ?? {});
}

function pathId(path: string): string {
  return decodeURIComponent(path.slice(path.lastIndexOf('/') + 1));
}

function bodyRequest(value: unknown): { body: unknown } {
  return { body: toOneKitJson(value ?? {}) };
}

function pathBody(path: string, value: unknown): { id: string; body: unknown } {
  return { id: pathId(path), body: toOneKitJson(value ?? {}) };
}

function dispatchGet(path: string, params: unknown): Promise<unknown> {
  const query = normalizeParams(params);
  if (path === '/hr/employees/me') return invoke('getMyEmployee', {});
  if (path === '/hr/staff/detail') return invoke('getStaffDetail', { id: String(query.id ?? '') });
  if (path === '/hr/time-off') return invoke('getTimeOff', {});
  if (path === '/hr/shifts/availability') return invoke('getAvailability', query);
  if (path === '/hr/shifts/roster') return invoke('getRoster', query);
  if (path === '/hr/shifts') return invoke('getShifts', query);
  if (path === '/hr/timesheets/current') return invoke('getCurrentTimesheet', {});
  if (path === '/hr/timesheets') return invoke('getTimesheets', query);
  if (path.startsWith('/hr/timesheets/')) return invoke('getTimesheet', { id: pathId(path) });
  if (path === '/hr/training') return invoke('getTraining', query);
  if (path === '/hr/documents/expiring') return invoke('getExpiringDocuments', query);
  if (path === '/hr/documents') return invoke('getDocuments', query);
  if (path === '/hr/benefits') return invoke('getBenefits', query);
  if (path === '/hr/staff') return invoke('getStaff', query);
  if (path === '/hr/config') return invoke('getConfig', {});
  if (path === '/hr/dashboard/v1/dashboard/widgets') return invoke('getDashboardWidgets', {});
  if (path === '/hr/dashboard/v1/dashboard') return invoke('getDashboard', {});
  if (path === '/hr/employees') return invoke('getEmployees', query);
  if (path.startsWith('/hr/employees/')) return invoke('getEmployee', { id: pathId(path) });
  if (path === '/hr/attendance') return invoke('getAttendance', query);
  if (path === '/hr/roles') return invoke('getRoles', {});
  if (path === '/hr/permissions') return invoke('getPermissions', {});
  if (path === '/hr/onboarding/templates') return invoke('getOnboardingTemplates', {});
  if (path === '/hr/onboarding/checklists') return invoke('getOnboardingChecklists', query);
  if (path === '/hr/onboarding/checklists/summary') return invoke('getOnboardingSummary', query);
  if (path === '/hr/performance/cycles') return invoke('getReviewCycles', query);
  if (path.startsWith('/hr/performance/cycles/'))
    return invoke('getReviewCycle', { id: pathId(path) });
  if (path === '/hr/performance/goals') return invoke('getGoals', query);
  if (path === '/hr/performance/reviews') return invoke('getPerformanceReviews', query);
  if (path === '/hr/leave/calendar') return invoke('getLeaveCalendar', query);
  if (path === '/hr/setup/status') return invoke('getSetupStatus', {});
  if (path === '/hr/overtime/config') return invoke('getOvertimeConfig', {});
  if (path === '/hr/payroll/config') return invoke('getPayrollConfig', {});
  if (path === '/hr/payroll/runs') return invoke('getPayrollRuns', query);
  if (path === '/hr/payroll/my-runs') return invoke('getMyPayrollRuns', query);
  if (path.startsWith('/hr/payroll/runs/')) return invoke('getPayrollRun', { id: pathId(path) });
  if (path.startsWith('/hr/payroll/my-runs/'))
    return invoke('getMyPayrollRun', { id: pathId(path) });
  if (path === '/hr/time-entries') return invoke('getTimeEntries', query);
  if (path === '/hr/reviews') return invoke('getReviews', {});
  if (path === '/hr/one-on-ones') return invoke('getOneOnOnes', {});
  throw new Error(`Unknown HR compatibility GET route: ${path}`);
}

function dispatchPost(path: string, data: unknown): Promise<unknown> {
  if (path === '/hr/time-off') return invoke('submitTimeOff', bodyRequest(data));
  if (path === '/hr/time-off/approve') return invoke('approveTimeOff', bodyRequest(data));
  if (path === '/hr/shifts/availability') return invoke('setAvailability', bodyRequest(data));
  if (path === '/hr/shifts/roster/assignments')
    return invoke('createShiftAssignment', bodyRequest(data));
  if (path === '/hr/shifts/roster/publish') return invoke('publishRoster', bodyRequest(data));
  if (path === '/hr/shifts') return invoke('upsertShift', bodyRequest(data));
  if (path.startsWith('/hr/timesheets/') && path.endsWith('/submit'))
    return invoke('submitTimesheet', pathBody(path.replace(/\/submit$/, ''), data));
  if (path.startsWith('/hr/timesheets/') && path.endsWith('/approve'))
    return invoke('approveTimesheet', pathBody(path.replace(/\/approve$/, ''), data));
  if (path.startsWith('/hr/timesheets/') && path.endsWith('/reject'))
    return invoke('rejectTimesheet', pathBody(path.replace(/\/reject$/, ''), data));
  if (path === '/hr/training') return invoke('upsertTraining', bodyRequest(data));
  if (path === '/hr/documents') return invoke('createDocument', bodyRequest(data));
  if (path === '/hr/staff') return invoke('upsertStaff', bodyRequest(data));
  if (path === '/hr/staff/clock-in') return invoke('clockIn', bodyRequest(data));
  if (path === '/hr/staff/clock-out') return invoke('clockOut', bodyRequest(data));
  if (path === '/hr/dashboard/v1/dashboard/widgets')
    return invoke('updateDashboardWidgets', bodyRequest(data));
  if (path === '/hr/reports/generate') return invoke('generateReport', bodyRequest(data));
  if (path === '/hr/employees') return invoke('createEmployee', bodyRequest(data));
  if (path === '/hr/employees/invite') return invoke('inviteEmployee', bodyRequest(data));
  if (path === '/hr/attendance/clock-in') return invoke('clockInAttendance', bodyRequest(data));
  if (path === '/hr/attendance/clock-out') return invoke('clockOutAttendance', bodyRequest(data));
  if (path === '/hr/roles') return invoke('createRole', bodyRequest(data));
  if (path === '/hr/onboarding/templates')
    return invoke('createOnboardingTemplate', bodyRequest(data));
  if (path.startsWith('/hr/onboarding/checklists/') && path.endsWith('/complete'))
    return invoke('completeOnboardingItem', pathBody(path.replace(/\/complete$/, ''), data));
  if (path === '/hr/performance/cycles') return invoke('createReviewCycle', bodyRequest(data));
  if (path.startsWith('/hr/performance/cycles/') && path.endsWith('/open'))
    return invoke('openReviewCycle', pathBody(path.replace(/\/open$/, ''), data));
  if (path === '/hr/performance/goals') return invoke('createGoal', bodyRequest(data));
  if (path.startsWith('/hr/performance/goals/')) return invoke('updateGoal', pathBody(path, data));
  if (path.startsWith('/hr/performance/reviews/') && path.endsWith('/submit'))
    return invoke('submitPerformanceReview', pathBody(path.replace(/\/submit$/, ''), data));
  if (path.startsWith('/hr/performance/reviews/') && path.endsWith('/share'))
    return invoke('sharePerformanceReview', pathBody(path.replace(/\/share$/, ''), data));
  if (path === '/hr/availability') return invoke('setAvailability', bodyRequest(data));
  if (path === '/hr/locations') return invoke('createLocation', bodyRequest(data));
  if (path.startsWith('/hr/setup/steps/') && path.endsWith('/complete'))
    return invoke('completeSetupStep', pathBody(path.replace(/\/complete$/, ''), data));
  if (path === '/hr/payroll/runs/trigger') return invoke('triggerPayrollRun', bodyRequest(data));
  if (path.startsWith('/hr/payroll/runs/') && path.endsWith('/approve'))
    return invoke('approvePayrollRun', pathBody(path.replace(/\/approve$/, ''), data));
  if (path === '/hr/time-entries') return invoke('createTimeEntry', bodyRequest(data));
  if (path === '/hr/time-entries/import-flow')
    return invoke('importTimeEntries', bodyRequest(data));
  if (path === '/hr/reviews') return invoke('createReview', bodyRequest(data));
  if (path === '/hr/one-on-ones') return invoke('upsertOneOnOne', bodyRequest(data));
  throw new Error(`Unknown HR compatibility POST route: ${path}`);
}

function dispatchPut(path: string, data: unknown): Promise<unknown> {
  if (path === '/hr/staff') return invoke('updateStaff', bodyRequest(data));
  if (path === '/hr/shifts') return invoke('updateShift', bodyRequest(data));
  if (path === '/hr/config') return invoke('updateConfig', bodyRequest(data));
  if (path === '/hr/dashboard/v1/dashboard/widgets')
    return invoke('updateDashboardWidgets', bodyRequest(data));
  if (path === '/hr/employees/me/pin') return invoke('updateEmployeePin', bodyRequest(data));
  if (path === '/hr/overtime/config') return invoke('updateOvertimeConfig', bodyRequest(data));
  if (path === '/hr/payroll/config') return invoke('updatePayrollConfig', bodyRequest(data));
  if (path.startsWith('/hr/employees/')) return invoke('updateEmployee', pathBody(path, data));
  if (path.startsWith('/hr/roles/')) return invoke('updateRole', pathBody(path, data));
  if (path.startsWith('/hr/performance/goals/')) return invoke('updateGoal', pathBody(path, data));
  if (path.startsWith('/hr/time-entries/')) return invoke('updateTimeEntry', pathBody(path, data));
  throw new Error(`Unknown HR compatibility PUT route: ${path}`);
}

function dispatchDelete(path: string, config?: HrRequestConfig): Promise<unknown> {
  const data = normalizeParams(config?.data);
  if (path === '/hr/staff') return staffClient.deleteStaff({ id: String(data.id ?? '') });
  if (path === '/hr/shifts') return shiftClient.deleteShift({ id: String(data.id ?? '') });
  if (path.startsWith('/hr/time-entries/')) return invoke('deleteTimeEntry', { id: pathId(path) });
  if (path.startsWith('/hr/documents/')) return invoke('deleteDocument', { id: pathId(path) });
  throw new Error(`Unknown HR compatibility DELETE route: ${path}`);
}

function hrResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((value) => fromOneKitJson<T>(value));
}

export const hrTransport = {
  get: <T = unknown>(path: string, config?: HrRequestConfig) =>
    hrResult<T>(() => dispatchGet(path, config?.params)),
  post: <T = unknown>(path: string, data?: unknown) => hrResult<T>(() => dispatchPost(path, data)),
  put: <T = unknown>(path: string, data?: unknown) => hrResult<T>(() => dispatchPut(path, data)),
  delete: <T = unknown>(path: string, config?: HrRequestConfig) =>
    hrResult<T>(() => dispatchDelete(path, config)),
};

export const licenseClient = createOneKitClient(LicenseServiceClient, 'hr');
export const userClient = createOneKitClient(UserServiceClient, 'hr');
