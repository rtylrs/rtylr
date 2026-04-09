/**
 * Local mock store for HR features that don't yet have backend endpoints.
 * Persists to localStorage so data survives reloads during development.
 */

type Listener = () => void;

const STORAGE_KEY = 'hr-mock-store-v1';

export interface EmployeeProfile {
  id: number;
  photoUrl?: string;
  phone?: string;
  address?: string;
  birthday?: string;   // YYYY-MM-DD
  hiredAt?: string;    // YYYY-MM-DD
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
  day: string;       // YYYY-MM-DD
  start: string;     // HH:mm
  end: string;       // HH:mm
  color?: string;
  published?: boolean;
}

export interface TimesheetRow {
  id: number;
  staffId: number;
  staffName: string;
  weekOf: string;    // YYYY-MM-DD (Monday)
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
  overall: number;     // 1-5
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
  scheduledAt: string;   // ISO
  duration: number;      // minutes
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
  progress: number; // 0-100
  completedAt?: string;
}

export interface Birthday {
  id: number;
  staffId: number;
  name: string;
  date: string; // MM-DD
  department?: string;
}

export interface Anniversary {
  id: number;
  staffId: number;
  name: string;
  date: string; // MM-DD
  years: number;
}

export interface Availability {
  id: number;
  staffId: number;
  staffName: string;
  day: string;      // e.g. Mon, Tue
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
  start: string;  // ISO
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

export interface MockStoreShape {
  profiles: EmployeeProfile[];
  timeOff: TimeOffEntry[];
  schedule: ScheduleShift[];
  timesheets: TimesheetRow[];
  reviews: Review[];
  oneOnOnes: OneOnOne[];
  onboarding: OnboardingTask[];
  offboarding: OffboardingTask[];
  documents: DocumentItem[];
  benefitPlans: BenefitPlan[];
  benefitEnrollments: BenefitEnrollment[];
  trainingCourses: TrainingCourse[];
  trainingProgress: TrainingProgress[];
  birthdays: Birthday[];
  anniversaries: Anniversary[];
  availability: Availability[];
  swaps: SwapRequest[];
  breaks: BreakLog[];
  overtimeAlerts: OvertimeAlert[];
  settings: HrSettings;
  nextId: number;
}

function seed(): MockStoreShape {
  return {
    profiles: [
      { id: 1, phone: '+1 415 555 0101', address: '100 Mission St, SF', birthday: '1990-04-22', hiredAt: '2021-03-15', emergency: { name: 'Jamie Chen', relation: 'Spouse', phone: '+1 415 555 0102' }, bio: 'Senior backend engineer who loves TypeScript and long walks in the terminal.', skills: ['Go', 'Postgres', 'Kubernetes'] },
      { id: 2, phone: '+1 415 555 0201', address: '250 Market St, SF', birthday: '1988-07-14', hiredAt: '2020-09-01', emergency: { name: 'Dana Reyes', relation: 'Sibling', phone: '+1 415 555 0202' }, bio: 'Product designer passionate about motion and micro interactions.', skills: ['Figma', 'Framer', 'UX research'] },
      { id: 3, phone: '+1 415 555 0301', address: '500 Howard St, SF', birthday: '1992-11-02', hiredAt: '2022-01-20', emergency: { name: 'Sam Park', relation: 'Parent', phone: '+1 415 555 0302' }, bio: 'Customer experience lead.', skills: ['Zendesk', 'Communication'] },
    ],
    timeOff: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', type: 'Vacation', startDate: '2026-04-15', endDate: '2026-04-19', reason: 'Family trip', status: 'Approved', createdAt: '2026-03-01' },
      { id: 2, staffId: 2, staffName: 'Morgan Reyes', type: 'Sick', startDate: '2026-04-10', endDate: '2026-04-10', reason: 'Flu', status: 'Pending', createdAt: '2026-04-09' },
      { id: 3, staffId: 3, staffName: 'Taylor Park', type: 'Personal', startDate: '2026-05-01', endDate: '2026-05-02', reason: 'Moving day', status: 'Pending', createdAt: '2026-04-05' },
    ],
    schedule: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', role: 'Engineer', day: '2026-04-13', start: '09:00', end: '17:00', color: '#8b5cf6', published: true },
      { id: 2, staffId: 2, staffName: 'Morgan Reyes', role: 'Designer', day: '2026-04-13', start: '10:00', end: '18:00', color: '#10b981', published: true },
      { id: 3, staffId: 3, staffName: 'Taylor Park', role: 'Support', day: '2026-04-14', start: '08:00', end: '16:00', color: '#3b82f6', published: true },
      { id: 4, staffId: 1, staffName: 'Alex Chen', role: 'Engineer', day: '2026-04-15', start: '09:00', end: '17:00', color: '#8b5cf6', published: true },
    ],
    timesheets: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', weekOf: '2026-04-06', hours: { mon: 8, tue: 8, wed: 8, thu: 8, fri: 7, sat: 0, sun: 0 }, approved: true },
      { id: 2, staffId: 2, staffName: 'Morgan Reyes', weekOf: '2026-04-06', hours: { mon: 7.5, tue: 8, wed: 8, thu: 8, fri: 8, sat: 0, sun: 0 }, approved: false },
      { id: 3, staffId: 3, staffName: 'Taylor Park', weekOf: '2026-04-06', hours: { mon: 8, tue: 8, wed: 8, thu: 8, fri: 8, sat: 4, sun: 0 }, approved: false },
    ],
    reviews: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', reviewerId: 2, reviewer: 'Morgan Reyes', period: 'Q1 2026', status: 'Completed', overall: 5, goals: 5, values: 4, skills: 5, notes: 'Exceeded goals on platform reliability.', date: '2026-03-28' },
      { id: 2, staffId: 2, staffName: 'Morgan Reyes', reviewerId: 1, reviewer: 'Alex Chen', period: 'Q1 2026', status: 'In Progress', overall: 4, goals: 4, values: 5, skills: 4, notes: 'Strong design leadership.', date: '2026-04-01' },
    ],
    oneOnOnes: [
      { id: 1, managerId: 1, manager: 'Alex Chen', reportId: 3, report: 'Taylor Park', scheduledAt: '2026-04-15T15:00:00Z', duration: 30, status: 'Scheduled', topics: ['Onboarding check-in', 'Career goals'] },
      { id: 2, managerId: 2, manager: 'Morgan Reyes', reportId: 1, report: 'Alex Chen', scheduledAt: '2026-04-10T10:00:00Z', duration: 45, status: 'Completed', topics: ['Q1 review', 'Roadmap'] , notes: 'Good discussion on roadmap priorities.' },
    ],
    onboarding: [
      { id: 1, staffId: 3, staffName: 'Taylor Park', task: 'Sign offer letter', category: 'Paperwork', completed: true, dueDate: '2026-04-01' },
      { id: 2, staffId: 3, staffName: 'Taylor Park', task: 'Laptop setup', category: 'Equipment', completed: true, dueDate: '2026-04-02' },
      { id: 3, staffId: 3, staffName: 'Taylor Park', task: 'Payroll enrollment', category: 'Paperwork', completed: false, dueDate: '2026-04-10' },
      { id: 4, staffId: 3, staffName: 'Taylor Park', task: 'Security training', category: 'Training', completed: false, dueDate: '2026-04-15' },
      { id: 5, staffId: 3, staffName: 'Taylor Park', task: 'Meet the team coffee', category: 'Meet the team', completed: false, dueDate: '2026-04-12' },
    ],
    offboarding: [
      { id: 1, staffId: 2, staffName: 'Morgan Reyes', task: 'Knowledge transfer doc', category: 'Knowledge transfer', completed: false },
      { id: 2, staffId: 2, staffName: 'Morgan Reyes', task: 'Revoke access tokens', category: 'Access', completed: false },
      { id: 3, staffId: 2, staffName: 'Morgan Reyes', task: 'Return laptop', category: 'Equipment', completed: false },
    ],
    documents: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', name: 'I-9 Form.pdf', type: 'I-9', uploadedAt: '2021-03-15', sizeKb: 210 },
      { id: 2, staffId: 1, staffName: 'Alex Chen', name: 'W4 2026.pdf', type: 'W4', uploadedAt: '2026-01-05', sizeKb: 145 },
      { id: 3, staffId: 2, staffName: 'Morgan Reyes', name: 'NDA.pdf', type: 'NDA', uploadedAt: '2020-09-01', expiresAt: '2026-05-15', sizeKb: 188 },
      { id: 4, staffId: 3, staffName: 'Taylor Park', name: 'Offer Letter.pdf', type: 'Offer Letter', uploadedAt: '2026-03-28', sizeKb: 320 },
    ],
    benefitPlans: [
      { id: 1, name: 'BlueShield PPO', category: 'Health', carrier: 'Blue Shield', monthlyCost: 480, enrolledCount: 12 },
      { id: 2, name: 'Delta Dental Premier', category: 'Dental', carrier: 'Delta', monthlyCost: 55, enrolledCount: 10 },
      { id: 3, name: 'Vision Plus', category: 'Vision', carrier: 'VSP', monthlyCost: 18, enrolledCount: 8 },
      { id: 4, name: 'Fidelity 401k Plan', category: '401k', carrier: 'Fidelity', monthlyCost: 0, enrolledCount: 15 },
    ],
    benefitEnrollments: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', planId: 1, planName: 'BlueShield PPO', startedAt: '2021-04-01', status: 'Active' },
      { id: 2, staffId: 1, staffName: 'Alex Chen', planId: 4, planName: 'Fidelity 401k Plan', startedAt: '2021-04-01', status: 'Active' },
      { id: 3, staffId: 2, staffName: 'Morgan Reyes', planId: 1, planName: 'BlueShield PPO', startedAt: '2020-10-01', status: 'Active' },
    ],
    trainingCourses: [
      { id: 1, title: 'Workplace Safety 101', category: 'Safety', durationMin: 45, modules: 6, required: true },
      { id: 2, title: 'Anti Harassment', category: 'Compliance', durationMin: 30, modules: 4, required: true },
      { id: 3, title: 'Leadership Basics', category: 'Leadership', durationMin: 90, modules: 8, required: false },
      { id: 4, title: 'React & TypeScript', category: 'Skills', durationMin: 240, modules: 12, required: false },
    ],
    trainingProgress: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', courseId: 1, courseName: 'Workplace Safety 101', progress: 100, completedAt: '2026-02-10' },
      { id: 2, staffId: 1, staffName: 'Alex Chen', courseId: 4, courseName: 'React & TypeScript', progress: 75 },
      { id: 3, staffId: 2, staffName: 'Morgan Reyes', courseId: 2, courseName: 'Anti Harassment', progress: 50 },
      { id: 4, staffId: 3, staffName: 'Taylor Park', courseId: 1, courseName: 'Workplace Safety 101', progress: 25 },
    ],
    birthdays: [
      { id: 1, staffId: 1, name: 'Alex Chen', date: '04-22', department: 'Engineering' },
      { id: 2, staffId: 2, name: 'Morgan Reyes', date: '07-14', department: 'Design' },
      { id: 3, staffId: 3, name: 'Taylor Park', date: '11-02', department: 'Support' },
    ],
    anniversaries: [
      { id: 1, staffId: 1, name: 'Alex Chen', date: '03-15', years: 5 },
      { id: 2, staffId: 2, name: 'Morgan Reyes', date: '09-01', years: 5 },
      { id: 3, staffId: 3, name: 'Taylor Park', date: '01-20', years: 4 },
    ],
    availability: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', day: 'Mon', canWork: true, startTime: '09:00', endTime: '17:00' },
      { id: 2, staffId: 1, staffName: 'Alex Chen', day: 'Tue', canWork: true, startTime: '09:00', endTime: '17:00' },
      { id: 3, staffId: 2, staffName: 'Morgan Reyes', day: 'Wed', canWork: false },
    ],
    swaps: [
      { id: 1, requesterId: 1, requester: 'Alex Chen', targetId: 2, target: 'Morgan Reyes', shiftDay: '2026-04-15', reason: 'Medical appointment', status: 'Pending' },
    ],
    breaks: [
      { id: 1, staffId: 1, staffName: 'Alex Chen', start: '2026-04-08T12:00:00Z', end: '2026-04-08T12:30:00Z', durationMin: 30, type: 'Lunch' },
    ],
    overtimeAlerts: [
      { id: 1, staffId: 3, staffName: 'Taylor Park', hoursThisWeek: 44, threshold: 40, triggeredAt: '2026-04-08' },
    ],
    settings: {
      workWeekStart: 'Monday',
      overtimeAfterHours: 40,
      ptoPolicyDays: 15,
      currency: 'USD',
      weeklyHours: 40,
    },
    nextId: 1000,
  };
}

function load(): MockStoreShape {
  if (typeof localStorage === 'undefined') return seed();
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      const fresh = seed();
      localStorage.setItem(STORAGE_KEY, JSON.stringify(fresh));
      return fresh;
    }
    return JSON.parse(raw);
  } catch {
    return seed();
  }
}

let state: MockStoreShape = load();
const listeners = new Set<Listener>();

function persist() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    // ignore
  }
  listeners.forEach((l) => l());
}

export const mockStore = {
  subscribe(fn: Listener) {
    listeners.add(fn);
    return () => listeners.delete(fn);
  },
  getState(): MockStoreShape {
    return state;
  },
  setState(updater: (s: MockStoreShape) => MockStoreShape) {
    state = updater(state);
    persist();
  },
  nextId() {
    state.nextId += 1;
    persist();
    return state.nextId;
  },
  reset() {
    state = seed();
    persist();
  },
};
