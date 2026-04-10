/**
 * Extended HR service layer backed by real API endpoints.
 */
import api from '../lib/api';

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
  hours: { mon: number; tue: number; wed: number; thu: number; fri: number; sat: number; sun: number };
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
}

// ─── API service ────────────────────────────────────────────────────────────

export const hrExtras = {
  // Profiles
  getProfile: (id: number) =>
    api.get<EmployeeProfile>(`/hr/staff/detail`, { params: { id } }).then((r) => r.data ?? null),
  upsertProfile: (p: EmployeeProfile) =>
    api.post<EmployeeProfile>('/hr/staff', p).then((r) => r.data),

  // Time Off
  listTimeOff: () =>
    api.get<TimeOffEntry[]>('/hr/time-off').then((r) => r.data ?? []),
  submitTimeOff: (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) =>
    api.post<TimeOffEntry>('/hr/time-off', data).then((r) => r.data),
  approveTimeOff: (id: number, approved: boolean, note?: string) =>
    api.post<{ success: boolean }>('/hr/time-off/approve', { id, approved, note }).then((r) => r.data),

  // Schedule
  listSchedule: (weekStart?: string) =>
    api.get<ScheduleShift[]>('/hr/shifts', { params: weekStart ? { weekStart } : {} }).then((r) => r.data ?? []),
  upsertScheduleShift: (shift: Omit<ScheduleShift, 'id'> & { id?: number }) =>
    api.post<{ success: boolean }>('/hr/shifts', shift).then((r) => r.data),
  deleteScheduleShift: (id: number) =>
    api.post<{ success: boolean }>('/hr/shifts', { id, _action: 'delete' }).then((r) => r.data),
  publishWeek: (weekStart: string) =>
    api.post<{ success: boolean }>('/hr/shifts', { weekStart, _action: 'publish' }).then((r) => r.data),

  // Timesheets
  listTimesheets: () =>
    api.get<TimesheetRow[]>('/hr/timesheets').then((r) => r.data ?? []),
  approveTimesheet: (id: number) =>
    api.post<{ success: boolean }>('/hr/timesheets', { id, _action: 'approve' }).then((r) => r.data),
  upsertTimesheet: (row: TimesheetRow) =>
    api.post<TimesheetRow>('/hr/timesheets', row).then((r) => r.data),

  // Reviews
  listReviews: () =>
    api.get<Review[]>('/hr/reviews').then((r) => r.data ?? []),
  createReview: (data: Omit<Review, 'id'>) =>
    api.post<Review>('/hr/reviews', data).then((r) => r.data),

  // 1-on-1s
  listOneOnOnes: () =>
    api.get<OneOnOne[]>('/hr/one-on-ones').then((r) => r.data ?? []),
  createOneOnOne: (data: Omit<OneOnOne, 'id'>) =>
    api.post<OneOnOne>('/hr/one-on-ones', data).then((r) => r.data),
  updateOneOnOne: (id: number, patch: Partial<OneOnOne>) =>
    api.post<{ success: boolean }>('/hr/one-on-ones', { id, ...patch }).then((r) => r.data),

  // Onboarding
  listOnboarding: () =>
    api.get<OnboardingTask[]>('/hr/training').then((r) => r.data ?? []),
  toggleOnboardingTask: (id: number) =>
    api.post<{ success: boolean }>('/hr/training', { id, _action: 'toggle' }).then((r) => r.data),
  addOnboardingTask: (task: Omit<OnboardingTask, 'id' | 'completed'>) =>
    api.post<OnboardingTask>('/hr/training', task).then((r) => r.data),

  // Offboarding
  listOffboarding: () =>
    api.get<OffboardingTask[]>('/hr/training', { params: { type: 'offboarding' } }).then((r) => r.data ?? []),
  toggleOffboardingTask: (id: number) =>
    api.post<{ success: boolean }>('/hr/training', { id, _action: 'toggle-offboarding' }).then((r) => r.data),

  // Documents
  listDocuments: () =>
    api.get<DocumentItem[]>('/hr/documents').then((r) => r.data ?? []),
  addDocument: (doc: Omit<DocumentItem, 'id' | 'uploadedAt'>) =>
    api.post<DocumentItem>('/hr/documents', doc).then((r) => r.data),
  deleteDocument: (id: number) =>
    api.post<{ success: boolean }>('/hr/documents', { id, _action: 'delete' }).then((r) => r.data),
  expiringDocuments: (daysAhead = 30) =>
    api.get<DocumentItem[]>('/hr/documents', { params: { expiring: daysAhead } }).then((r) => r.data ?? []),

  // Benefits
  listBenefitPlans: () =>
    api.get<BenefitPlan[]>('/hr/benefits').then((r) => r.data ?? []),
  listBenefitEnrollments: () =>
    api.get<BenefitEnrollment[]>('/hr/benefits', { params: { type: 'enrollments' } }).then((r) => r.data ?? []),
  enrollBenefit: (data: Omit<BenefitEnrollment, 'id' | 'startedAt' | 'status'>) =>
    api.post<BenefitEnrollment>('/hr/benefits', data).then((r) => r.data),

  // Training
  listCourses: () =>
    api.get<TrainingCourse[]>('/hr/training', { params: { type: 'courses' } }).then((r) => r.data ?? []),
  listTrainingProgress: () =>
    api.get<TrainingProgress[]>('/hr/training', { params: { type: 'progress' } }).then((r) => r.data ?? []),
  updateTrainingProgress: (id: number, progress: number) =>
    api.post<{ success: boolean }>('/hr/training', { id, progress, _action: 'update-progress' }).then((r) => r.data),

  // Birthdays & anniversaries
  listBirthdays: () =>
    api.get<Birthday[]>('/hr/staff', { params: { view: 'birthdays' } }).then((r) => r.data ?? []),
  listAnniversaries: () =>
    api.get<Anniversary[]>('/hr/staff', { params: { view: 'anniversaries' } }).then((r) => r.data ?? []),

  // Availability
  listAvailability: () =>
    api.get<Availability[]>('/hr/shifts', { params: { view: 'availability' } }).then((r) => r.data ?? []),
  setAvailability: (row: Availability) =>
    api.post<Availability>('/hr/shifts', { ...row, _action: 'set-availability' }).then((r) => r.data),

  // Swaps
  listSwaps: () =>
    api.get<SwapRequest[]>('/hr/shifts', { params: { view: 'swaps' } }).then((r) => r.data ?? []),
  createSwap: (data: Omit<SwapRequest, 'id' | 'status'>) =>
    api.post<SwapRequest>('/hr/shifts', { ...data, _action: 'create-swap' }).then((r) => r.data),
  resolveSwap: (id: number, accepted: boolean) =>
    api.post<{ success: boolean }>('/hr/shifts', { id, accepted, _action: 'resolve-swap' }).then((r) => r.data),

  // Breaks
  listBreaks: () =>
    api.get<BreakLog[]>('/hr/shifts', { params: { view: 'breaks' } }).then((r) => r.data ?? []),
  startBreak: (staffId: number, staffName: string, type: BreakLog['type']) =>
    api.post<BreakLog>('/hr/shifts/clock-in', { staffId, staffName, type, _action: 'start-break' }).then((r) => r.data),
  endBreak: (id: number) =>
    api.post<{ success: boolean }>('/hr/shifts/clock-out', { id, _action: 'end-break' }).then((r) => r.data),

  // Overtime
  listOvertimeAlerts: () =>
    api.get<OvertimeAlert[]>('/hr/shifts', { params: { view: 'overtime' } }).then((r) => r.data ?? []),

  // Settings
  getSettings: () =>
    api.get<HrSettings>('/hr/dashboard', { params: { view: 'settings' } }).then((r) => r.data ?? {
      workWeekStart: 'Monday' as const,
      overtimeAfterHours: 40,
      ptoPolicyDays: 15,
      currency: 'USD' as const,
      weeklyHours: 40,
    }),
  updateSettings: (patch: Partial<HrSettings>) =>
    api.post<HrSettings>('/hr/dashboard', { ...patch, _action: 'update-settings' }).then((r) => r.data),
};
