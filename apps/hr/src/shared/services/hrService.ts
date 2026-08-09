import { safe } from '../lib/api';

export interface StaffPayload {
  name: string;
  email?: string;
  phone?: string;
  role?: string;
  departmentId?: number;
  startDate?: string;
  salary?: number;
  [key: string]: unknown;
}

export interface DepartmentPayload {
  name: string;
  managerId?: number;
  [key: string]: unknown;
}

export interface ShiftPayload {
  staffId: number;
  start: string;
  end: string;
  locationId?: number;
  [key: string]: unknown;
}

export interface ClockPayload {
  staffId?: number;
  locationId?: number;
  [key: string]: unknown;
}

export interface TimeOffPayload {
  staffId: number;
  startDate: string;
  endDate: string;
  type?: string;
  reason?: string;
  [key: string]: unknown;
}

export interface TimeOffApprovalPayload {
  id: number;
  status: 'approved' | 'rejected';
  notes?: string;
  [key: string]: unknown;
}

export interface PayrollPayload {
  period: string;
  staffIds?: number[];
  [key: string]: unknown;
}

export interface MarkPaidPayload {
  payrollId: number;
  paidAt?: string;
  [key: string]: unknown;
}

export const hrService = {
  // Staff
  getStaff: () => safe.get('/hr/staff'),
  createStaff: (data: StaffPayload) => safe.post('/hr/staff', data),
  updateStaff: (id: number, data: Partial<StaffPayload>) => safe.put('/hr/staff', { id, ...data }),
  deleteStaff: (id: number) => safe.delete('/hr/staff', { data: { id } }),
  getDashboard: () => safe.get('/hr/dashboard/v1/dashboard'),

  // Departments
  getDepartments: () => safe.get('/hr/departments'),
  createDepartment: (data: DepartmentPayload) => safe.post('/hr/departments', data),
  updateDepartment: (id: number, data: Partial<DepartmentPayload>) =>
    safe.put('/hr/departments', { id, ...data }),

  // Shifts
  getShifts: () => safe.get('/hr/shifts'),
  createShift: (data: ShiftPayload) => safe.post('/hr/shifts', data),
  clockIn: (data: ClockPayload) => safe.post('/hr/staff/clock-in', data),
  clockOut: (data: ClockPayload) => safe.post('/hr/staff/clock-out', data),

  // Time Off
  getTimeOff: () => safe.get('/hr/time-off'),
  submitTimeOff: (data: TimeOffPayload) => safe.post('/hr/time-off', data),
  approveTimeOff: (data: TimeOffApprovalPayload) => safe.post('/hr/time-off/approve', data),

  // Payroll
  getPayroll: () => safe.get('/hr/payroll'),
  runPayroll: (data: PayrollPayload) => safe.post('/hr/payroll/run', data),
  markPayrollPaid: (data: MarkPaidPayload) => safe.post('/hr/payroll/mark-paid', data),
};
