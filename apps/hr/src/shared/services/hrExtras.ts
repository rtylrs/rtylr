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

const toProfilePayload = (p: EmployeeProfile) => ({
  id: p.id,
  phone: p.phone,
  address: p.address,
  emergencyContactName: p.emergency?.name,
  emergencyContactPhone: p.emergency?.phone,
  notes: p.bio,
});

const toTimeOffPayload = (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) => ({
  staffId: data.staffId,
  type: data.type,
  startDate: data.startDate,
  endDate: data.endDate,
  reason: data.reason,
  notes: data.note,
});

const toScheduleShiftCreatePayload = (shift: Omit<ScheduleShift, 'id'> & { id?: number }) => ({
  staffId: shift.staffId,
  date: shift.day,
  startTime: shift.start,
  endTime: shift.end,
  status: shift.published ? 'published' : undefined,
});

const toScheduleShiftUpdatePayload = (shift: Omit<ScheduleShift, 'id'> & { id?: number }) => ({
  id: shift.id,
  date: shift.day,
  startTime: shift.start,
  endTime: shift.end,
  status: shift.published ? 'published' : undefined,
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

const toBenefitPayload = (data: Omit<BenefitEnrollment, 'id' | 'startedAt' | 'status'>) => ({
  name: data.planName,
  type: 'enrollment',
  provider: '',
  costEmployer: 0,
  costEmployee: 0,
  isActive: true,
});

const toSettingsPayload = (patch: Partial<HrSettings>) => ({
  setupCompletedAt:
    'setupCompletedAt' in patch
      ? (patch as { setupCompletedAt?: number }).setupCompletedAt
      : undefined,
});

// ─── API service ────────────────────────────────────────────────────────────

export const hrExtras = {
  // Profiles
  getProfile: (id: number) => safe.get<EmployeeProfile>(`/hr/staff/detail`, { params: { id } }),
  upsertProfile: (p: EmployeeProfile) =>
    safe.put<EmployeeProfile>('/hr/staff', toProfilePayload(p)),

  // Time Off
  listTimeOff: () => safe.get<TimeOffEntry[]>('/hr/time-off'),
  submitTimeOff: (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) =>
    safe.post<TimeOffEntry>('/hr/time-off', toTimeOffPayload(data)),
  approveTimeOff: (id: number, approved: boolean, note?: string) =>
    safe.post<{ success: boolean }>('/hr/time-off/approve', {
      id,
      status: approved ? 'approved' : 'rejected',
      notes: note,
    }),

  // Schedule
  listSchedule: (weekStart?: string) =>
    safe.get<ScheduleShift[]>('/hr/shifts', { params: weekStart ? { weekStart } : {} }),
  upsertScheduleShift: (shift: Omit<ScheduleShift, 'id'> & { id?: number }) =>
    shift.id
      ? safe.put<{ success: boolean }>('/hr/shifts', toScheduleShiftUpdatePayload(shift))
      : safe.post<{ success: boolean }>('/hr/shifts', toScheduleShiftCreatePayload(shift)),
  deleteScheduleShift: (id: number) =>
    safe.delete<{ success: boolean }>('/hr/shifts', { data: { id } }),
  publishWeek: (weekStart: string) =>
    safe.get<{ success: boolean }>('/hr/shifts', { params: { weekStart, view: 'publish' } }),

  // Timesheets
  listTimesheets: () => safe.get<TimesheetRow[]>('/hr/timesheets'),
  approveTimesheet: (id: number) =>
    safe.post<{ success: boolean }>(`/hr/timesheets/${id}/approve`, {}),
  upsertTimesheet: (row: TimesheetRow) =>
    safe.post<TimesheetRow>(`/hr/timesheets/${row.id}/submit`, {}),

  // Reviews
  listReviews: () => safe.get<Review[]>('/hr/reviews'),
  createReview: (data: Omit<Review, 'id'>) =>
    safe.post<Review>('/hr/reviews', toReviewPayload(data)),

  // 1-on-1s
  listOneOnOnes: () => safe.get<OneOnOne[]>('/hr/one-on-ones'),
  createOneOnOne: (data: Omit<OneOnOne, 'id'>) =>
    safe.post<OneOnOne>('/hr/one-on-ones', toOneOnOnePayload(data)),
  updateOneOnOne: (id: number, patch: Partial<OneOnOne>) =>
    safe.post<{ success: boolean }>('/hr/one-on-ones', { id, ...patch }),

  // Onboarding
  listOnboarding: () => safe.get<OnboardingTask[]>('/hr/training'),
  toggleOnboardingTask: (id: number) =>
    safe.get<{ success: boolean }>('/hr/training', { params: { id, action: 'toggle' } }),
  addOnboardingTask: (task: Omit<OnboardingTask, 'id' | 'completed'>) =>
    safe.post<OnboardingTask>('/hr/training', toTrainingPayload(task)),

  // Offboarding
  listOffboarding: () =>
    safe.get<OffboardingTask[]>('/hr/training', { params: { type: 'offboarding' } }),
  toggleOffboardingTask: (id: number) =>
    safe.get<{ success: boolean }>('/hr/training', {
      params: { id, action: 'toggle-offboarding' },
    }),

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
    safe.post<BenefitEnrollment>('/hr/benefits', toBenefitPayload(data)),

  // Training
  listCourses: () => safe.get<TrainingCourse[]>('/hr/training', { params: { type: 'courses' } }),
  listTrainingProgress: () =>
    safe.get<TrainingProgress[]>('/hr/training', { params: { type: 'progress' } }),
  updateTrainingProgress: (id: number, progress: number) =>
    safe.get<{ success: boolean }>('/hr/training', {
      params: { id, progress, action: 'update-progress' },
    }),

  // Birthdays & anniversaries
  listBirthdays: () => safe.get<Birthday[]>('/hr/staff', { params: { view: 'birthdays' } }),
  listAnniversaries: () =>
    safe.get<Anniversary[]>('/hr/staff', { params: { view: 'anniversaries' } }),

  // Availability
  listAvailability: () =>
    safe.get<Availability[]>('/hr/shifts', { params: { view: 'availability' } }),
  setAvailability: (row: Availability) =>
    safe.put<Availability>('/hr/shifts', {
      id: row.id,
      date: row.day,
      startTime: row.startTime,
      endTime: row.endTime,
    }),

  // Swaps
  listSwaps: () => safe.get<SwapRequest[]>('/hr/shifts', { params: { view: 'swaps' } }),
  createSwap: (data: Omit<SwapRequest, 'id' | 'status'>) =>
    safe.post<SwapRequest>('/hr/shifts', {
      staffId: data.requesterId,
      date: data.shiftDay,
      notes: data.reason,
    }),
  resolveSwap: (id: number, accepted: boolean) =>
    safe.put<{ success: boolean }>('/hr/shifts', {
      id,
      status: accepted ? 'accepted' : 'rejected',
    }),

  // Breaks
  listBreaks: () => safe.get<BreakLog[]>('/hr/shifts', { params: { view: 'breaks' } }),
  startBreak: (staffId: number, staffName: string, type: BreakLog['type']) =>
    safe.post<BreakLog>('/hr/staff/clock-in', {
      staffId,
    }),
  endBreak: (id: number) => safe.post<{ success: boolean }>('/hr/staff/clock-out', { staffId: id }),

  // Overtime
  listOvertimeAlerts: () =>
    safe.get<OvertimeAlert[]>('/hr/shifts', { params: { view: 'overtime' } }),

  // Settings
  getSettings: () => safe.get<HrSettings>('/hr/dashboard', { params: { view: 'settings' } }),
  updateSettings: (patch: Partial<HrSettings>) =>
    safe.put<HrSettings>('/hr/config', toSettingsPayload(patch)),
};
