import { safe } from '../lib/api';

// ─── Types ───────────────────────────────────────────────────────────────────

export interface StaffMember {
  id: number;
  name: string;
  role: string;
  department: string;
  email: string;
  employmentType: 'Full-time' | 'Part-time' | 'Contractor';
  status: 'Active' | 'Inactive';
}

export interface CreateStaffInput {
  name: string;
  role: string;
  department: string;
  email: string;
  employmentType: 'Full-time' | 'Part-time' | 'Contractor';
}

export interface UpdateStaffInput extends Partial<CreateStaffInput> {
  id: number;
}

export interface Department {
  id: number;
  name: string;
  description: string;
  staffCount: number;
  head: string;
  avgSalary: number;
  color?: string;
}

export interface CreateDepartmentInput {
  name: string;
  description: string;
  head: string;
}

export interface UpdateDepartmentInput extends Partial<CreateDepartmentInput> {
  id: number;
}

export interface Shift {
  id: number;
  name: string;
  role: string;
  start: string;
  end: string;
  status: 'Scheduled' | 'In Progress' | 'Completed';
  day: string;
  staffId?: number;
}

export interface CreateShiftInput {
  staffId: number;
  start: string;
  end: string;
  day: string;
}

export interface UpdateShiftInput extends Partial<CreateShiftInput> {
  id: number;
}

export interface PayrollRow {
  id: number;
  name: string;
  role: string;
  hoursWorked: number;
  basePay: number;
  overtime: number;
  deductions: number;
  status?: 'Pending' | 'Paid';
}

export interface TimeOffRequest {
  id: number;
  staffId: number;
  staffName: string;
  startDate: string;
  endDate: string;
  reason: string;
  status: 'Pending' | 'Approved' | 'Rejected';
}

export interface HRDashboardStats {
  totalStaff: number;
  onShiftToday: number;
  openPositions: number;
  monthlyPayroll: number;
  recentActivity: Array<{
    name: string;
    action: string;
    time: string;
    type: 'in' | 'out';
    role: string;
  }>;
  todayAtGlance: {
    shiftsScheduled: number;
    newApplications: number;
    pendingApprovals: number;
  };
}

const splitName = (name?: string) => {
  const [firstName = '', ...rest] = (name ?? '').trim().split(/\s+/).filter(Boolean);
  return { firstName, lastName: rest.join(' ') || firstName };
};

const toStaffPayload = (data: Partial<CreateStaffInput> & { id?: number }) => {
  const { firstName, lastName } = splitName(data.name);
  return {
    firstName,
    lastName,
    email: data.email,
    role: data.role,
    employmentType: data.employmentType,
  };
};

const toStaffUpdatePayload = (data: UpdateStaffInput) => ({
  id: data.id,
  ...toStaffPayload(data),
});

const toDepartmentCreatePayload = (data: CreateDepartmentInput) => ({
  name: data.name,
  description: data.description,
});

const toDepartmentUpdatePayload = (data: UpdateDepartmentInput) => ({
  id: data.id,
  name: data.name,
  description: data.description,
});

const toShiftCreatePayload = (data: CreateShiftInput) => ({
  staffId: data.staffId,
  date: data.day,
  startTime: data.start,
  endTime: data.end,
});

const toShiftUpdatePayload = (data: UpdateShiftInput) => ({
  id: data.id,
  date: data.day,
  startTime: data.start,
  endTime: data.end,
});

// ─── API ─────────────────────────────────────────────────────────────────────

export const hrApi = {
  // Staff
  listStaff: () => safe.get<StaffMember[]>('/hr/staff'),
  createStaff: (data: CreateStaffInput) =>
    safe.post<StaffMember>('/hr/staff', toStaffPayload(data)),
  updateStaff: (data: UpdateStaffInput) =>
    safe.put<StaffMember>('/hr/staff', toStaffUpdatePayload(data)),
  deleteStaff: (id: number) => safe.delete<{ success: boolean }>('/hr/staff', { data: { id } }),
  clockIn: (staffId: number) =>
    safe.post<{ success: boolean; time: string }>('/hr/staff/clock-in', { staffId }),
  clockOut: (staffId: number) =>
    safe.post<{ success: boolean; time: string }>('/hr/staff/clock-out', { staffId }),

  // Departments
  listDepartments: () => safe.get<Department[]>('/hr/departments'),
  createDepartment: (data: CreateDepartmentInput) =>
    safe.post<Department>('/hr/departments', toDepartmentCreatePayload(data)),
  updateDepartment: (data: UpdateDepartmentInput) =>
    safe.put<Department>('/hr/departments', toDepartmentUpdatePayload(data)),

  // Shifts
  listShifts: (params?: { day?: string; staffId?: number }) =>
    safe.get<Shift[]>('/hr/shifts', { params }),
  createShift: (data: CreateShiftInput) =>
    safe.post<Shift>('/hr/shifts', toShiftCreatePayload(data)),
  updateShift: (data: UpdateShiftInput) =>
    safe.put<Shift>('/hr/shifts', toShiftUpdatePayload(data)),

  // Payroll
  listPayroll: (params?: { period?: string }) => safe.get<PayrollRow[]>('/hr/payroll', { params }),
  runPayroll: (period: string) =>
    safe.post<{ success: boolean }>('/hr/payroll/run', {
      periodStart: period,
      periodEnd: period,
      periodLabel: period,
    }),
  markPayrollPaid: (id: number) => safe.post<{ success: boolean }>('/hr/payroll/mark-paid', { id }),

  // Time Off
  listTimeOff: () => safe.get<TimeOffRequest[]>('/hr/time-off'),
  submitTimeOff: (data: { staffId: number; startDate: string; endDate: string; reason: string }) =>
    safe.post<TimeOffRequest>('/hr/time-off', data),
  approveTimeOff: (id: number, approved: boolean) =>
    safe.post<{ success: boolean }>('/hr/time-off/approve', {
      id,
      status: approved ? 'approved' : 'rejected',
    }),

  // Dashboard
  getDashboard: () => safe.get<HRDashboardStats>('/hr/dashboard'),
  getDashboardWidgets: () => safe.get<unknown>('/hr/dashboard/widgets'),
  updateWidgetLayout: (widgets: unknown[]) =>
    safe.put<{ success: boolean }>('/hr/dashboard/widgets', { widgets }),

  // Reports
  generateReport: (request: unknown) => safe.post<unknown>('/hr/reports/generate', { request }),
};
