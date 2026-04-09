/**
 * Extended HR service layer backed by the local mock store.
 * Provides promise-based APIs to integrate cleanly with TanStack Query.
 */
import {
  mockStore,
  EmployeeProfile,
  TimeOffEntry,
  ScheduleShift,
  TimesheetRow,
  Review,
  OneOnOne,
  OnboardingTask,
  OffboardingTask,
  DocumentItem,
  BenefitPlan,
  BenefitEnrollment,
  TrainingCourse,
  TrainingProgress,
  Birthday,
  Anniversary,
  Availability,
  SwapRequest,
  BreakLog,
  OvertimeAlert,
  HrSettings,
} from './mockStore';

function asyncOk<T>(value: T, delay = 120): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), delay));
}

export const hrExtras = {
  // Profiles
  getProfile: (id: number) => asyncOk(mockStore.getState().profiles.find((p) => p.id === id) ?? null),
  upsertProfile: (p: EmployeeProfile) => {
    mockStore.setState((s) => {
      const idx = s.profiles.findIndex((x) => x.id === p.id);
      if (idx >= 0) s.profiles[idx] = { ...s.profiles[idx], ...p };
      else s.profiles.push(p);
      return s;
    });
    return asyncOk(p);
  },

  // Time Off
  listTimeOff: () => asyncOk([...mockStore.getState().timeOff]),
  submitTimeOff: (data: Omit<TimeOffEntry, 'id' | 'status' | 'createdAt'>) => {
    const id = mockStore.nextId();
    const created: TimeOffEntry = {
      ...data,
      id,
      status: 'Pending',
      createdAt: new Date().toISOString().slice(0, 10),
    };
    mockStore.setState((s) => ({ ...s, timeOff: [created, ...s.timeOff] }));
    return asyncOk(created);
  },
  approveTimeOff: (id: number, approved: boolean, note?: string) => {
    mockStore.setState((s) => ({
      ...s,
      timeOff: s.timeOff.map((t) => (t.id === id ? { ...t, status: approved ? 'Approved' : 'Rejected', note } : t)),
    }));
    return asyncOk({ success: true });
  },

  // Schedule
  listSchedule: (weekStart?: string) => {
    const all = mockStore.getState().schedule;
    if (!weekStart) return asyncOk(all);
    const start = new Date(weekStart);
    const end = new Date(weekStart);
    end.setDate(end.getDate() + 7);
    return asyncOk(all.filter((s) => {
      const d = new Date(s.day);
      return d >= start && d < end;
    }));
  },
  upsertScheduleShift: (shift: Omit<ScheduleShift, 'id'> & { id?: number }) => {
    mockStore.setState((s) => {
      if (shift.id) {
        return { ...s, schedule: s.schedule.map((x) => (x.id === shift.id ? { ...x, ...shift } as ScheduleShift : x)) };
      }
      const id = mockStore.nextId();
      return { ...s, schedule: [...s.schedule, { ...shift, id } as ScheduleShift] };
    });
    return asyncOk({ success: true });
  },
  deleteScheduleShift: (id: number) => {
    mockStore.setState((s) => ({ ...s, schedule: s.schedule.filter((x) => x.id !== id) }));
    return asyncOk({ success: true });
  },
  publishWeek: (weekStart: string) => {
    mockStore.setState((s) => ({
      ...s,
      schedule: s.schedule.map((sh) => {
        const d = new Date(sh.day);
        const ws = new Date(weekStart);
        const we = new Date(weekStart);
        we.setDate(we.getDate() + 7);
        return d >= ws && d < we ? { ...sh, published: true } : sh;
      }),
    }));
    return asyncOk({ success: true });
  },

  // Timesheets
  listTimesheets: () => asyncOk([...mockStore.getState().timesheets]),
  approveTimesheet: (id: number) => {
    mockStore.setState((s) => ({
      ...s,
      timesheets: s.timesheets.map((t) => (t.id === id ? { ...t, approved: true } : t)),
    }));
    return asyncOk({ success: true });
  },
  upsertTimesheet: (row: TimesheetRow) => {
    mockStore.setState((s) => {
      const idx = s.timesheets.findIndex((t) => t.id === row.id);
      if (idx >= 0) s.timesheets[idx] = row;
      else s.timesheets.push(row);
      return s;
    });
    return asyncOk(row);
  },

  // Reviews
  listReviews: () => asyncOk([...mockStore.getState().reviews]),
  createReview: (data: Omit<Review, 'id'>) => {
    const id = mockStore.nextId();
    const created = { ...data, id };
    mockStore.setState((s) => ({ ...s, reviews: [created, ...s.reviews] }));
    return asyncOk(created);
  },

  // 1-on-1s
  listOneOnOnes: () => asyncOk([...mockStore.getState().oneOnOnes]),
  createOneOnOne: (data: Omit<OneOnOne, 'id'>) => {
    const id = mockStore.nextId();
    const created = { ...data, id };
    mockStore.setState((s) => ({ ...s, oneOnOnes: [created, ...s.oneOnOnes] }));
    return asyncOk(created);
  },
  updateOneOnOne: (id: number, patch: Partial<OneOnOne>) => {
    mockStore.setState((s) => ({
      ...s,
      oneOnOnes: s.oneOnOnes.map((o) => (o.id === id ? { ...o, ...patch } : o)),
    }));
    return asyncOk({ success: true });
  },

  // Onboarding
  listOnboarding: () => asyncOk([...mockStore.getState().onboarding]),
  toggleOnboardingTask: (id: number) => {
    mockStore.setState((s) => ({
      ...s,
      onboarding: s.onboarding.map((o) => (o.id === id ? { ...o, completed: !o.completed } : o)),
    }));
    return asyncOk({ success: true });
  },
  addOnboardingTask: (task: Omit<OnboardingTask, 'id' | 'completed'>) => {
    const id = mockStore.nextId();
    const created: OnboardingTask = { ...task, id, completed: false };
    mockStore.setState((s) => ({ ...s, onboarding: [...s.onboarding, created] }));
    return asyncOk(created);
  },

  // Offboarding
  listOffboarding: () => asyncOk([...mockStore.getState().offboarding]),
  toggleOffboardingTask: (id: number) => {
    mockStore.setState((s) => ({
      ...s,
      offboarding: s.offboarding.map((o) => (o.id === id ? { ...o, completed: !o.completed } : o)),
    }));
    return asyncOk({ success: true });
  },

  // Documents
  listDocuments: () => asyncOk([...mockStore.getState().documents]),
  addDocument: (doc: Omit<DocumentItem, 'id' | 'uploadedAt'>) => {
    const id = mockStore.nextId();
    const created: DocumentItem = { ...doc, id, uploadedAt: new Date().toISOString().slice(0, 10) };
    mockStore.setState((s) => ({ ...s, documents: [created, ...s.documents] }));
    return asyncOk(created);
  },
  deleteDocument: (id: number) => {
    mockStore.setState((s) => ({ ...s, documents: s.documents.filter((d) => d.id !== id) }));
    return asyncOk({ success: true });
  },
  expiringDocuments: (daysAhead = 30) => {
    const now = new Date();
    const cutoff = new Date();
    cutoff.setDate(now.getDate() + daysAhead);
    return asyncOk(
      mockStore.getState().documents.filter((d) => {
        if (!d.expiresAt) return false;
        const expires = new Date(d.expiresAt);
        return expires >= now && expires <= cutoff;
      }),
    );
  },

  // Benefits
  listBenefitPlans: () => asyncOk([...mockStore.getState().benefitPlans]),
  listBenefitEnrollments: () => asyncOk([...mockStore.getState().benefitEnrollments]),
  enrollBenefit: (data: Omit<BenefitEnrollment, 'id' | 'startedAt' | 'status'>) => {
    const id = mockStore.nextId();
    const created: BenefitEnrollment = { ...data, id, startedAt: new Date().toISOString().slice(0, 10), status: 'Active' };
    mockStore.setState((s) => ({ ...s, benefitEnrollments: [created, ...s.benefitEnrollments] }));
    return asyncOk(created);
  },

  // Training
  listCourses: () => asyncOk([...mockStore.getState().trainingCourses]),
  listTrainingProgress: () => asyncOk([...mockStore.getState().trainingProgress]),
  updateTrainingProgress: (id: number, progress: number) => {
    mockStore.setState((s) => ({
      ...s,
      trainingProgress: s.trainingProgress.map((t) =>
        t.id === id ? { ...t, progress, completedAt: progress >= 100 ? new Date().toISOString().slice(0, 10) : undefined } : t,
      ),
    }));
    return asyncOk({ success: true });
  },

  // Birthdays & anniversaries
  listBirthdays: () => asyncOk([...mockStore.getState().birthdays]),
  listAnniversaries: () => asyncOk([...mockStore.getState().anniversaries]),

  // Availability
  listAvailability: () => asyncOk([...mockStore.getState().availability]),
  setAvailability: (row: Availability) => {
    mockStore.setState((s) => {
      const idx = s.availability.findIndex((a) => a.staffId === row.staffId && a.day === row.day);
      if (idx >= 0) s.availability[idx] = row;
      else s.availability.push(row);
      return s;
    });
    return asyncOk(row);
  },

  // Swaps
  listSwaps: () => asyncOk([...mockStore.getState().swaps]),
  createSwap: (data: Omit<SwapRequest, 'id' | 'status'>) => {
    const id = mockStore.nextId();
    const created: SwapRequest = { ...data, id, status: 'Pending' };
    mockStore.setState((s) => ({ ...s, swaps: [created, ...s.swaps] }));
    return asyncOk(created);
  },
  resolveSwap: (id: number, accepted: boolean) => {
    mockStore.setState((s) => ({
      ...s,
      swaps: s.swaps.map((x) => (x.id === id ? { ...x, status: accepted ? 'Accepted' : 'Rejected' } : x)),
    }));
    return asyncOk({ success: true });
  },

  // Breaks
  listBreaks: () => asyncOk([...mockStore.getState().breaks]),
  startBreak: (staffId: number, staffName: string, type: BreakLog['type']) => {
    const id = mockStore.nextId();
    const created: BreakLog = { id, staffId, staffName, start: new Date().toISOString(), type };
    mockStore.setState((s) => ({ ...s, breaks: [created, ...s.breaks] }));
    return asyncOk(created);
  },
  endBreak: (id: number) => {
    mockStore.setState((s) => ({
      ...s,
      breaks: s.breaks.map((b) => {
        if (b.id !== id || b.end) return b;
        const end = new Date().toISOString();
        const durationMin = Math.round((new Date(end).getTime() - new Date(b.start).getTime()) / 60000);
        return { ...b, end, durationMin };
      }),
    }));
    return asyncOk({ success: true });
  },

  // Overtime
  listOvertimeAlerts: () => asyncOk([...mockStore.getState().overtimeAlerts]),

  // Settings
  getSettings: () => asyncOk({ ...mockStore.getState().settings }),
  updateSettings: (patch: Partial<HrSettings>) => {
    mockStore.setState((s) => ({ ...s, settings: { ...s.settings, ...patch } }));
    return asyncOk(mockStore.getState().settings);
  },
};

export type {
  EmployeeProfile,
  TimeOffEntry,
  ScheduleShift,
  TimesheetRow,
  Review,
  OneOnOne,
  OnboardingTask,
  OffboardingTask,
  DocumentItem,
  BenefitPlan,
  BenefitEnrollment,
  TrainingCourse,
  TrainingProgress,
  Birthday,
  Anniversary,
  Availability,
  SwapRequest,
  BreakLog,
  OvertimeAlert,
  HrSettings,
};
