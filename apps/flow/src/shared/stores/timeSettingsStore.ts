import { create } from 'zustand';
import { persist } from 'zustand/middleware';

/**
 * Workspace time settings.
 *
 * These are workspace values with shipped defaults, not product constants: a law
 * firm approving weekly through a partner and a restaurant where the manager
 * clears everything on Sunday need different defaults, not different code.
 *
 * There is no workspace time-settings endpoint yet. This store is the seam that
 * keeps that fact in one file - every screen reads from here, so landing the
 * endpoint replaces this store's internals and touches nothing else. Until then
 * the values are per-browser, and the settings screen says so.
 *
 * See docs/flow-time-tracking.md for the full gap list.
 */

export type WeekStart = 'monday' | 'sunday';

/** 0 means no rounding. */
export type RoundingMinutes = 0 | 5 | 10 | 15 | 30;

interface TimeCategory {
  id: string;
  name: string;
  billable: boolean;
}

interface TimeSettingsState {
  weekStart: WeekStart;
  roundingMinutes: RoundingMinutes;
  weeklyTargetMinutes: number;
  overtimeThresholdMinutes: number;
  /** On: an approved entry is read-only until a manager unapproves it. */
  approvalLocksEntry: boolean;
  /** Entries auto-lock this many days after the week closes, either way. */
  autoLockDays: number;
  categories: TimeCategory[];
  /** userId to hourly rate. `null` is explicitly unset, never zero. */
  ratesByUserId: Record<string, number | null>;
}

interface TimeSettingsActions {
  setWeekStart: (value: WeekStart) => void;
  setRoundingMinutes: (value: RoundingMinutes) => void;
  setWeeklyTargetMinutes: (value: number) => void;
  setOvertimeThresholdMinutes: (value: number) => void;
  setApprovalLocksEntry: (value: boolean) => void;
  setAutoLockDays: (value: number) => void;
  addCategory: (name: string, billable: boolean) => void;
  renameCategory: (id: string, name: string) => void;
  setCategoryBillable: (id: string, billable: boolean) => void;
  deleteCategory: (id: string) => void;
  setRate: (userId: string, rate: number | null) => void;
}

export const TIME_SETTINGS_STORAGE_KEY = 'flow.timeSettings-v1';

/**
 * The four seeded categories are a starting point, not a taxonomy - a workspace
 * is expected to rewrite them. Admin defaults to non-billable because internal
 * admin time is the one of the four that is usually absorbed.
 */
export const DEFAULT_TIME_SETTINGS: TimeSettingsState = {
  weekStart: 'monday',
  roundingMinutes: 15,
  weeklyTargetMinutes: 40 * 60,
  overtimeThresholdMinutes: 45 * 60,
  approvalLocksEntry: true,
  autoLockDays: 7,
  categories: [
    { id: 'development', name: 'Development', billable: true },
    { id: 'review', name: 'Review', billable: true },
    { id: 'meeting', name: 'Meeting', billable: true },
    { id: 'admin', name: 'Admin', billable: false },
  ],
  ratesByUserId: {},
};

function slugify(name: string): string {
  return name.trim().toLowerCase().replace(/\s+/g, '-');
}

export const useTimeSettingsStore = create<TimeSettingsState & TimeSettingsActions>()(
  persist(
    (set, get) => ({
      ...DEFAULT_TIME_SETTINGS,

      setWeekStart: (weekStart) => set({ weekStart }),
      setRoundingMinutes: (roundingMinutes) => set({ roundingMinutes }),
      setWeeklyTargetMinutes: (weeklyTargetMinutes) => set({ weeklyTargetMinutes }),
      setOvertimeThresholdMinutes: (overtimeThresholdMinutes) => set({ overtimeThresholdMinutes }),
      setApprovalLocksEntry: (approvalLocksEntry) => set({ approvalLocksEntry }),
      setAutoLockDays: (autoLockDays) => set({ autoLockDays }),

      addCategory: (name, billable) => {
        const trimmed = name.trim();
        if (!trimmed) return;
        const existing = get().categories;
        if (existing.some((c) => c.name.toLowerCase() === trimmed.toLowerCase())) return;
        // Suffix the slug on collision so two categories renamed into the same
        // shape cannot end up sharing an id.
        let id = slugify(trimmed);
        if (existing.some((c) => c.id === id)) id = `${id}-${existing.length + 1}`;
        set({ categories: [...existing, { id, name: trimmed, billable }] });
      },

      renameCategory: (id, name) => {
        const trimmed = name.trim();
        if (!trimmed) return;
        set({
          categories: get().categories.map((c) => (c.id === id ? { ...c, name: trimmed } : c)),
        });
      },

      setCategoryBillable: (id, billable) =>
        set({
          categories: get().categories.map((c) => (c.id === id ? { ...c, billable } : c)),
        }),

      deleteCategory: (id) => {
        const remaining = get().categories.filter((c) => c.id !== id);
        // An empty list would leave an entry with no category to pick.
        if (remaining.length === 0) return;
        set({ categories: remaining });
      },

      setRate: (userId, rate) => set({ ratesByUserId: { ...get().ratesByUserId, [userId]: rate } }),
    }),
    { name: TIME_SETTINGS_STORAGE_KEY }
  )
);

/**
 * Applies the workspace rounding increment to a logged duration.
 *
 * Rounds to the nearest increment, except that a real entry never rounds away to
 * nothing: 5 minutes under a 15-minute increment becomes 15, not 0. Nearest
 * would silently discard the shortest entries, which is the one rounding error
 * that loses work rather than shifting it.
 */
export function roundLoggedMinutes(minutes: number, increment: RoundingMinutes): number {
  if (increment === 0 || minutes <= 0) return minutes;
  const rounded = Math.round(minutes / increment) * increment;
  return rounded === 0 ? increment : rounded;
}
