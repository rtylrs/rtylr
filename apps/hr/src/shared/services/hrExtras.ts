/**
 * Extended HR service layer backed by real API endpoints.
 */
import { hrTransport } from '../lib/clients';

// ─── Types ──────────────────────────────────────────────────────────────────

export interface TimeOffEntry {
  id: number;
  staffId: number;
  staffName: string;
  type: 'Vacation' | 'Sick' | 'Personal' | 'Bereavement' | 'Unpaid';
  startDate: string;
  endDate: string;
  reason: string;
  status: 'Pending' | 'Approved' | 'Rejected';
  note?: string;
  createdAt: string;
}

export interface ScheduleShift {
  id: number;
  staffId: number;
  staffName: string;
  role: string;
  day: string;
  start: string;
  end: string;
  color?: string;
  published?: boolean;
}

export interface TimesheetRow {
  id: number;
  staffId: number;
  staffName: string;
  weekOf: string;
  hours: {
    mon: number;
    tue: number;
    wed: number;
    thu: number;
    fri: number;
    sat: number;
    sun: number;
  };
  approved?: boolean;
}

export interface Review {
  id: number;
  staffId: number;
  staffName: string;
  reviewerId: number;
  reviewer: string;
  period: string;
  status: 'Draft' | 'In Progress' | 'Submitted' | 'Completed';
  overall: number;
  goals: number;
  values: number;
  skills: number;
  notes: string;
  date: string;
}

export interface OneOnOne {
  id: number;
  managerId: number;
  manager: string;
  reportId: number;
  report: string;
  scheduledAt: string;
  duration: number;
  status: 'Scheduled' | 'Completed' | 'Cancelled';
  topics: string[];
  notes?: string;
}

export interface OnboardingTask {
  id: number;
  staffId: number;
  staffName: string;
  task: string;
  category: 'Paperwork' | 'Access' | 'Training' | 'Meet the team' | 'Equipment';
  completed: boolean;
  dueDate?: string;
}

export interface OffboardingTask {
  id: number;
  staffId: number;
  staffName: string;
  task: string;
  category: 'Knowledge transfer' | 'Access' | 'Equipment' | 'Paperwork';
  completed: boolean;
  dueDate?: string;
}

export interface DocumentItem {
  id: number;
  staffId: number;
  staffName: string;
  name: string;
  type: 'I-9' | 'W4' | 'Contract' | 'NDA' | 'Offer Letter' | 'Other';
  uploadedAt: string;
  expiresAt?: string;
  sizeKb: number;
  url?: string;
}

export interface BenefitPlan {
  id: number;
  name: string;
  category: 'Health' | 'Dental' | 'Vision' | '401k' | 'Life';
  carrier: string;
  monthlyCost: number;
  enrolledCount: number;
}

export interface BenefitEnrollment {
  id: number;
  staffId: number;
  staffName: string;
  planId: number;
  planName: string;
  startedAt: string;
  status: 'Active' | 'Pending' | 'Waived';
}

export interface TrainingCourse {
  id: number;
  title: string;
  category: 'Compliance' | 'Safety' | 'Skills' | 'Leadership';
  durationMin: number;
  modules: number;
  required: boolean;
}

export interface TrainingProgress {
  id: number;
  staffId: number;
  staffName: string;
  courseId: number;
  courseName: string;
  progress: number;
  completedAt?: string;
}

export interface Birthday {
  id: number;
  staffId: number;
  name: string;
  date: string;
  department?: string;
}

export interface Anniversary {
  id: number;
  staffId: number;
  name: string;
  date: string;
  years: number;
}

export interface SwapRequest {
  id: number;
  requesterId: number;
  requester: string;
  targetId: number;
  target: string;
  shiftDay: string;
  reason: string;
  status: 'Pending' | 'Accepted' | 'Rejected';
}

export interface BreakLog {
  id: number;
  staffId: number;
  staffName: string;
  start: string;
  end?: string;
  durationMin?: number;
  type: 'Lunch' | 'Short' | 'Other';
}

export interface OvertimeAlert {
  id: number;
  staffId: number;
  staffName: string;
  hoursThisWeek: number;
  threshold: number;
  triggeredAt: string;
}

export interface HrSettings {
  workWeekStart: 'Monday' | 'Sunday';
  overtimeAfterHours: number;
  ptoPolicyDays: number;
  currency: 'USD' | 'EUR' | 'GBP';
  weeklyHours: number;
  industry?: string;
}

const toTimeOffPayload = (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) => ({
  staffId: data.staffId,
  type: data.type,
  startDate: data.startDate,
  endDate: data.endDate,
  reason: data.reason,
  notes: data.note,
});

// Keys are snake_case to match the API's request binding. "published" is not a
// shift status -- the status column only accepts scheduled, in_progress,
// completed, absent or cancelled, and publishing is a property of a roster
// assignment, not of the shift row -- so it is no longer sent here. Use
// publishWeek to publish a roster.
const toScheduleShiftCreatePayload = (shift: Omit<ScheduleShift, 'id'> & { id?: number }) => ({
  staff_id: shift.staffId,
  date: shift.day,
  start_time: shift.start,
  end_time: shift.end,
});

const toScheduleShiftUpdatePayload = (shift: Omit<ScheduleShift, 'id'> & { id?: number }) => ({
  id: shift.id,
  date: shift.day,
  start_time: shift.start,
  end_time: shift.end,
});

const toReviewPayload = (data: Omit<Review, 'id'>) => ({
  staffId: data.staffId,
  reviewerId: data.reviewerId,
  period: data.period,
  rating: data.overall,
  goalsRating: data.goals,
  valuesRating: data.values,
  skillsRating: data.skills,
  notes: data.notes,
  status: data.status,
});

const toOneOnOnePayload = (data: Omit<OneOnOne, 'id'>) => ({
  managerId: data.managerId,
  staffId: data.reportId,
  scheduledAt: data.scheduledAt,
  notes: data.notes,
  actionItems: data.topics.join('\n'),
  status: data.status,
});

const toTrainingPayload = (task: Omit<OnboardingTask, 'id' | 'completed'>) => ({
  name: task.task,
  description: task.category,
  durationHours: 0,
  isRequired: true,
});

const toSettingsPayload = (patch: Partial<HrSettings>): Partial<HrSettings> => ({
  workWeekStart: patch.workWeekStart,
  overtimeAfterHours: patch.overtimeAfterHours,
  ptoPolicyDays: patch.ptoPolicyDays,
  currency: patch.currency,
  weeklyHours: patch.weeklyHours,
  industry: patch.industry,
});

// ─── API service ────────────────────────────────────────────────────────────

export const hrExtras = {
  // My own employee record
  // hr.employee.v1 has no self-lookup RPC, and hr_employees.id is a different
  // id-space from the hub auth user id - this is the only correct way for the
  // frontend to resolve "my own employee_id" for leave/overtime/equipment forms.
  getMyEmployee: () => hrTransport.get<{ employee: { id: number | string } }>('/hr/employees/me'),

  // Time Off
  listTimeOff: () => hrTransport.get<TimeOffEntry[]>('/hr/time-off'),
  submitTimeOff: (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) =>
    hrTransport.post<TimeOffEntry>('/hr/time-off', toTimeOffPayload(data)),
  approveTimeOff: (id: number, approved: boolean, note?: string) =>
    hrTransport.post<{ success: boolean }>('/hr/time-off/approve', {
      id,
      status: approved ? 'approved' : 'rejected',
      notes: note,
    }),

  // Schedule
  listSchedule: (weekStart?: string) =>
    hrTransport.get<ScheduleShift[]>('/hr/shifts', {
      params: weekStart ? { date: weekStart } : {},
    }),
  upsertScheduleShift: (shift: Omit<ScheduleShift, 'id'> & { id?: number }) =>
    shift.id
      ? hrTransport.put<{ success: boolean }>('/hr/shifts', toScheduleShiftUpdatePayload(shift))
      : hrTransport.post<{ success: boolean }>('/hr/shifts', toScheduleShiftCreatePayload(shift)),
  deleteScheduleShift: (id: number) =>
    hrTransport.delete<{ success: boolean }>('/hr/shifts', { data: { id } }),
  // Publishing a roster is a POST to its own endpoint. This used to GET
  // /hr/shifts with view=publish, which no handler ever read, so it returned a
  // shift list and published nothing.
  publishWeek: (weekStart: string) =>
    hrTransport.post<{ success: boolean }>('/hr/shifts/roster/publish', { week_start: weekStart }),

  // Timesheets
  listTimesheets: () => hrTransport.get<TimesheetRow[]>('/hr/timesheets'),
  approveTimesheet: (id: number) =>
    hrTransport.post<{ success: boolean }>(`/hr/timesheets/${id}/approve`, {}),
  upsertTimesheet: (row: TimesheetRow) =>
    hrTransport.post<TimesheetRow>(`/hr/timesheets/${row.id}/submit`, {}),

  // Reviews
  listReviews: () => hrTransport.get<Review[]>('/hr/reviews'),
  createReview: (data: Omit<Review, 'id'>) =>
    hrTransport.post<Review>('/hr/reviews', toReviewPayload(data)),

  // 1-on-1s
  listOneOnOnes: () => hrTransport.get<OneOnOne[]>('/hr/one-on-ones'),
  createOneOnOne: (data: Omit<OneOnOne, 'id'>) =>
    hrTransport.post<OneOnOne>('/hr/one-on-ones', toOneOnOnePayload(data)),
  updateOneOnOne: (id: number, patch: Partial<OneOnOne>) =>
    hrTransport.post<{ success: boolean }>('/hr/one-on-ones', { id, ...patch }),

  // Onboarding
  listOnboarding: () => hrTransport.get<OnboardingTask[]>('/hr/training'),
  toggleOnboardingTask: (id: number) =>
    hrTransport.get<{ success: boolean }>('/hr/training', { params: { id, action: 'toggle' } }),
  addOnboardingTask: (task: Omit<OnboardingTask, 'id' | 'completed'>) =>
    hrTransport.post<OnboardingTask>('/hr/training', toTrainingPayload(task)),

  // Offboarding
  listOffboarding: () =>
    hrTransport.get<OffboardingTask[]>('/hr/training', { params: { type: 'offboarding' } }),
  toggleOffboardingTask: (id: number) =>
    hrTransport.get<{ success: boolean }>('/hr/training', {
      params: { id, action: 'toggle-offboarding' },
    }),

  // Documents
  listDocuments: () => hrTransport.get<DocumentItem[]>('/hr/documents'),
  addDocument: (doc: Omit<DocumentItem, 'id' | 'uploadedAt'>) =>
    hrTransport.post<DocumentItem>('/hr/documents', {
      employeeId: doc.staffId,
      documentType: doc.type,
      title: doc.name,
      storageMode: doc.url ? 'linked' : 'stored',
      fileUrl: doc.url ?? doc.name,
    }),
  deleteDocument: (id: number) => hrTransport.delete<{ success: boolean }>(`/hr/documents/${id}`),
  expiringDocuments: (daysAhead = 30) =>
    hrTransport.get<DocumentItem[]>('/hr/documents/expiring', { params: { days: daysAhead } }),

  // Benefits
  listBenefitPlans: () => hrTransport.get<BenefitPlan[]>('/hr/benefits'),
  // Legacy, per-employee admin view used by EmployeeProfile.tsx. Self-service
  // enrollment now goes through hr.benefit.v1 (see useMyBenefitEnrollments /
  // useEnrollInBenefit in useHrExtras.ts) - there is no onekit RPC yet for an
  // admin to list a specific employee's enrollments, so this stays on the
  // legacy REST path until one exists.
  listBenefitEnrollments: () =>
    hrTransport.get<BenefitEnrollment[]>('/hr/benefits', { params: { type: 'enrollments' } }),

  // Training
  listCourses: () =>
    hrTransport.get<TrainingCourse[]>('/hr/training', { params: { type: 'courses' } }),
  listTrainingProgress: () =>
    hrTransport.get<TrainingProgress[]>('/hr/training', { params: { type: 'progress' } }),
  updateTrainingProgress: (id: number, progress: number) =>
    hrTransport.get<{ success: boolean }>('/hr/training', {
      params: { id, progress, action: 'update-progress' },
    }),

  // Birthdays & anniversaries
  listBirthdays: () => hrTransport.get<Birthday[]>('/hr/staff', { params: { view: 'birthdays' } }),
  listAnniversaries: () =>
    hrTransport.get<Anniversary[]>('/hr/staff', { params: { view: 'anniversaries' } }),

  // Availability lives in shared/hooks/queries/useAvailability.ts, which calls
  // the hr.shift.v1 availability endpoints through the generated client. The
  // calls that used to sit here issued GET and PUT against /hr/shifts with a
  // view=availability parameter no handler has ever read.

  // Swaps
  listSwaps: () => hrTransport.get<SwapRequest[]>('/hr/shifts', { params: { view: 'swaps' } }),
  createSwap: (data: Omit<SwapRequest, 'id' | 'status'>) =>
    hrTransport.post<SwapRequest>('/hr/shifts', {
      staffId: data.requesterId,
      date: data.shiftDay,
      notes: data.reason,
    }),
  resolveSwap: (id: number, accepted: boolean) =>
    hrTransport.put<{ success: boolean }>('/hr/shifts', {
      id,
      status: accepted ? 'accepted' : 'rejected',
    }),

  // Breaks
  listBreaks: () => hrTransport.get<BreakLog[]>('/hr/shifts', { params: { view: 'breaks' } }),
  startBreak: (staffId: number, _staffName: string, _type: BreakLog['type']) =>
    hrTransport.post<BreakLog>('/hr/staff/clock-in', {
      staffId,
    }),
  endBreak: (id: number) =>
    hrTransport.post<{ success: boolean }>('/hr/staff/clock-out', { staffId: id }),

  // Overtime
  listOvertimeAlerts: () =>
    hrTransport.get<OvertimeAlert[]>('/hr/shifts', { params: { view: 'overtime' } }),

  // Settings
  getSettings: () => hrTransport.get<{ config: HrSettings }>('/hr/config'),
  updateSettings: (patch: Partial<HrSettings>) =>
    hrTransport.put<HrSettings>('/hr/config', toSettingsPayload(patch)),
};
