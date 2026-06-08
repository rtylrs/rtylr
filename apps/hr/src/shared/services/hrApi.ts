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

// ─── API ─────────────────────────────────────────────────────────────────────

export const hrApi = {
  // Staff
  listStaff: () => safe.get<StaffMember[]>('/hr/staff'),
  createStaff: (data: CreateStaffInput) => safe.post<StaffMember>('/hr/staff', data),
  updateStaff: (data: UpdateStaffInput) => safe.post<StaffMember>('/hr/staff', data),
  deleteStaff: (id: number) =>
    safe.post<{ success: boolean }>('/hr/staff', { id, _action: 'delete' }),
  clockIn: (staffId: number) =>
    safe.post<{ success: boolean; time: string }>('/hr/staff/clock-in', { staffId }),
  clockOut: (staffId: number) =>
    safe.post<{ success: boolean; time: string }>('/hr/staff/clock-out', { staffId }),

  // Departments
  listDepartments: () => safe.get<Department[]>('/hr/departments'),
  createDepartment: (data: CreateDepartmentInput) => safe.post<Department>('/hr/departments', data),
  updateDepartment: (data: UpdateDepartmentInput) => safe.post<Department>('/hr/departments', data),

  // Shifts
  listShifts: (params?: { day?: string; staffId?: number }) =>
    safe.get<Shift[]>('/hr/shifts', { params }),
  createShift: (data: CreateShiftInput) => safe.post<Shift>('/hr/shifts', data),
  updateShift: (data: UpdateShiftInput) => safe.post<Shift>('/hr/shifts', data),

  // Payroll
  listPayroll: (params?: { period?: string }) => safe.get<PayrollRow[]>('/hr/payroll', { params }),
  runPayroll: (period: string) => safe.post<{ success: boolean }>('/hr/payroll/run', { period }),
  markPayrollPaid: (id: number) => safe.post<{ success: boolean }>('/hr/payroll/mark-paid', { id }),

  // Time Off
  listTimeOff: () => safe.get<TimeOffRequest[]>('/hr/time-off'),
  submitTimeOff: (data: { staffId: number; startDate: string; endDate: string; reason: string }) =>
    safe.post<TimeOffRequest>('/hr/time-off', data),
  approveTimeOff: (id: number, approved: boolean) =>
    safe.post<{ success: boolean }>('/hr/time-off/approve', { id, approved }),

  // Dashboard
  getDashboard: () => safe.get<HRDashboardStats>('/hr/dashboard'),
  getDashboardWidgets: () => safe.get<unknown>('/hr/dashboard/widgets'),
  updateWidgetLayout: (widgets: unknown[]) =>
    safe.put<{ success: boolean }>('/hr/dashboard/widgets', { widgets }),

  // Reports
  generateReport: (request: unknown) => safe.post<unknown>('/hr/reports/generate', { request }),
};
