import { hrTransport } from '../lib/clients';

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

// The API binds request fields by their snake_case JSON tag. Its decoder tries
// the tag, then a lower-first variant, then an all-lower variant, so camelCase
// keys like startTime match nothing and arrive as zero values.
const toShiftCreatePayload = (data: CreateShiftInput) => ({
  staff_id: data.staffId,
  date: data.day,
  start_time: data.start,
  end_time: data.end,
});

const toShiftUpdatePayload = (data: UpdateShiftInput) => ({
  id: data.id,
  date: data.day,
  start_time: data.start,
  end_time: data.end,
});

// ─── API ─────────────────────────────────────────────────────────────────────

export const hrApi = {
  // Staff
  createStaff: (data: CreateStaffInput) =>
    hrTransport.post<StaffMember>('/hr/staff', toStaffPayload(data)),
  updateStaff: (data: UpdateStaffInput) =>
    hrTransport.put<StaffMember>('/hr/staff', toStaffUpdatePayload(data)),
  deleteStaff: (id: number) =>
    hrTransport.delete<{ success: boolean }>('/hr/staff', { data: { id } }),
  clockIn: (staffId: number) =>
    hrTransport.post<{ success: boolean; time: string }>('/hr/staff/clock-in', { staffId }),
  clockOut: (staffId: number) =>
    hrTransport.post<{ success: boolean; time: string }>('/hr/staff/clock-out', { staffId }),

  // Shifts
  listShifts: (params?: { day?: string; staffId?: number }) =>
    hrTransport.get<Shift[]>('/hr/shifts', {
      params: {
        ...(params?.day ? { date: params.day } : {}),
        ...(params?.staffId ? { staff_id: params.staffId } : {}),
      },
    }),
  createShift: (data: CreateShiftInput) =>
    hrTransport.post<Shift>('/hr/shifts', toShiftCreatePayload(data)),
  updateShift: (data: UpdateShiftInput) =>
    hrTransport.put<Shift>('/hr/shifts', toShiftUpdatePayload(data)),

  // Payroll

  // Dashboard
  getDashboard: () => hrTransport.get<HRDashboardStats>('/hr/dashboard/v1/dashboard'),
  getDashboardWidgets: () => hrTransport.get<unknown>('/hr/dashboard/v1/dashboard/widgets'),
  updateWidgetLayout: (widgets: unknown[]) =>
    hrTransport.put<{ success: boolean }>('/hr/dashboard/v1/dashboard/widgets', { widgets }),

  // Reports
  generateReport: (request: unknown) =>
    hrTransport.post<unknown>('/hr/reports/generate', { request }),
};
