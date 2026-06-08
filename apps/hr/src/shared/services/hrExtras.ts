/**
 * Extended HR service layer backed by real API endpoints.
 */
import { safe } from '../lib/api';
import { unwrap } from '@voxire/api';

// ─── Types ──────────────────────────────────────────────────────────────────

export interface EmployeeProfile {
  id: number;
  photoUrl?: string;
  phone?: string;
  address?: string;
  birthday?: string;
  hiredAt?: string;
  emergency?: { name: string; relation: string; phone: string };
  bio?: string;
  managerId?: number;
  skills?: string[];
}

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

export interface Availability {
  id: number;
  staffId: number;
  staffName: string;
  day: string;
  canWork: boolean;
  startTime?: string;
  endTime?: string;
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

// ─── API service ────────────────────────────────────────────────────────────

export const hrExtras = {
  // Profiles
  getProfile: (id: number) => safe.get<EmployeeProfile>(`/hr/staff/detail`, { params: { id } }),
  upsertProfile: (p: EmployeeProfile) => safe.post<EmployeeProfile>('/hr/staff', p),

  // Time Off
  listTimeOff: () => safe.get<TimeOffEntry[]>('/hr/time-off'),
  submitTimeOff: (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) =>
    safe.post<TimeOffEntry>('/hr/time-off', data),
  approveTimeOff: (id: number, approved: boolean, note?: string) =>
    safe.post<{ success: boolean }>('/hr/time-off/approve', { id, approved, note }),

  // Schedule
  listSchedule: (weekStart?: string) =>
    safe.get<ScheduleShift[]>('/hr/shifts', { params: weekStart ? { weekStart } : {} }),
  upsertScheduleShift: (shift: Omit<ScheduleShift, 'id'> & { id?: number }) =>
    safe.post<{ success: boolean }>('/hr/shifts', shift),
  deleteScheduleShift: (id: number) =>
    safe.post<{ success: boolean }>('/hr/shifts', { id, _action: 'delete' }),
  publishWeek: (weekStart: string) =>
    safe.post<{ success: boolean }>('/hr/shifts', { weekStart, _action: 'publish' }),

  // Timesheets
  listTimesheets: () => safe.get<TimesheetRow[]>('/hr/timesheets'),
  approveTimesheet: (id: number) =>
    safe.post<{ success: boolean }>('/hr/timesheets', { id, _action: 'approve' }),
  upsertTimesheet: (row: TimesheetRow) => safe.post<TimesheetRow>('/hr/timesheets', row),

  // Reviews
  listReviews: () => safe.get<Review[]>('/hr/reviews'),
  createReview: (data: Omit<Review, 'id'>) => safe.post<Review>('/hr/reviews', data),

  // 1-on-1s
  listOneOnOnes: () => safe.get<OneOnOne[]>('/hr/one-on-ones'),
  createOneOnOne: (data: Omit<OneOnOne, 'id'>) => safe.post<OneOnOne>('/hr/one-on-ones', data),
  updateOneOnOne: (id: number, patch: Partial<OneOnOne>) =>
    safe.post<{ success: boolean }>('/hr/one-on-ones', { id, ...patch }),

  // Onboarding
  listOnboarding: () => safe.get<OnboardingTask[]>('/hr/training'),
  toggleOnboardingTask: (id: number) =>
    safe.post<{ success: boolean }>('/hr/training', { id, _action: 'toggle' }),
  addOnboardingTask: (task: Omit<OnboardingTask, 'id' | 'completed'>) =>
    safe.post<OnboardingTask>('/hr/training', task),

  // Offboarding
  listOffboarding: () =>
    safe.get<OffboardingTask[]>('/hr/training', { params: { type: 'offboarding' } }),
  toggleOffboardingTask: (id: number) =>
    safe.post<{ success: boolean }>('/hr/training', { id, _action: 'toggle-offboarding' }),

  // Documents
  listDocuments: () => safe.get<DocumentItem[]>('/hr/documents'),
  addDocument: (doc: Omit<DocumentItem, 'id' | 'uploadedAt'>) =>
    safe.post<DocumentItem>('/hr/documents', {
      employeeId: doc.staffId,
      documentType: doc.type,
      title: doc.name,
      storageMode: doc.url ? 'linked' : 'stored',
      fileUrl: doc.url ?? doc.name,
    }),
  deleteDocument: (id: number) => safe.delete<{ success: boolean }>(`/hr/documents/${id}`),
  expiringDocuments: (daysAhead = 30) =>
    safe.get<DocumentItem[]>('/hr/documents/expiring', { params: { days: daysAhead } }),

  // Benefits
  listBenefitPlans: () => safe.get<BenefitPlan[]>('/hr/benefits'),
  listBenefitEnrollments: () =>
    safe.get<BenefitEnrollment[]>('/hr/benefits', { params: { type: 'enrollments' } }),
  enrollBenefit: (data: Omit<BenefitEnrollment, 'id' | 'startedAt' | 'status'>) =>
    safe.post<BenefitEnrollment>('/hr/benefits', data),

  // Training
  listCourses: () => safe.get<TrainingCourse[]>('/hr/training', { params: { type: 'courses' } }),
  listTrainingProgress: () =>
    safe.get<TrainingProgress[]>('/hr/training', { params: { type: 'progress' } }),
  updateTrainingProgress: (id: number, progress: number) =>
    safe.post<{ success: boolean }>('/hr/training', { id, progress, _action: 'update-progress' }),

  // Birthdays & anniversaries
  listBirthdays: () => safe.get<Birthday[]>('/hr/staff', { params: { view: 'birthdays' } }),
  listAnniversaries: () =>
    safe.get<Anniversary[]>('/hr/staff', { params: { view: 'anniversaries' } }),

  // Availability
  listAvailability: () =>
    safe.get<Availability[]>('/hr/shifts', { params: { view: 'availability' } }),
  setAvailability: (row: Availability) =>
    safe.post<Availability>('/hr/shifts', { ...row, _action: 'set-availability' }),

  // Swaps
  listSwaps: () => safe.get<SwapRequest[]>('/hr/shifts', { params: { view: 'swaps' } }),
  createSwap: (data: Omit<SwapRequest, 'id' | 'status'>) =>
    safe.post<SwapRequest>('/hr/shifts', { ...data, _action: 'create-swap' }),
  resolveSwap: (id: number, accepted: boolean) =>
    safe.post<{ success: boolean }>('/hr/shifts', { id, accepted, _action: 'resolve-swap' }),

  // Breaks
  listBreaks: () => safe.get<BreakLog[]>('/hr/shifts', { params: { view: 'breaks' } }),
  startBreak: (staffId: number, staffName: string, type: BreakLog['type']) =>
    safe.post<BreakLog>('/hr/shifts/clock-in', {
      staffId,
      staffName,
      type,
      _action: 'start-break',
    }),
  endBreak: (id: number) =>
    safe.post<{ success: boolean }>('/hr/shifts/clock-out', { id, _action: 'end-break' }),

  // Overtime
  listOvertimeAlerts: () =>
    safe.get<OvertimeAlert[]>('/hr/shifts', { params: { view: 'overtime' } }),

  // Settings
  getSettings: () => safe.get<HrSettings>('/hr/dashboard', { params: { view: 'settings' } }),
  updateSettings: (patch: Partial<HrSettings>) =>
    safe.post<HrSettings>('/hr/dashboard', { ...patch, _action: 'update-settings' }),
};
