import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { DAY_KEYS, type DayKey } from '../lib/dateUtils';

// These display and entry preferences persist in this browser.
export type WeekStart = 'monday' | 'sunday';

/** 0 means no rounding. */
export type RoundingMinutes = 0 | 5 | 10 | 15 | 30;

interface TimeSettingsState {
  weekStart: WeekStart;
  roundingMinutes: RoundingMinutes;
  /** Expected work minutes per calendar day - fully independent per day, not a
   * weekly figure split evenly, because a 4-day week or a short Saturday is a
   * real shape a business has, not an edge case of a uniform one. */
  dailyTargetMinutes: Record<DayKey, number>;
  overtimeThresholdMinutes: number;
  /** On: My Day and reports show Clocked/Unaccounted (shift attendance) alongside
   * Logged. Off: task-based logging only. Different businesses want different
   * defaults here - a restaurant needs shift reconciliation, a law firm billing
   * by task usually doesn't. */
  trackAttendance: boolean;
}

interface TimeSettingsActions {
  setWeekStart: (value: WeekStart) => void;
  setRoundingMinutes: (value: RoundingMinutes) => void;
  setDailyTargetMinutes: (day: DayKey, minutes: number) => void;
  setOvertimeThresholdMinutes: (value: number) => void;
  setTrackAttendance: (value: boolean) => void;
}

export const TIME_SETTINGS_STORAGE_KEY = 'flow.timeSettings-v1';

/** Returns the sum of every day's target - the "weekly target" a flat single
 * number used to be, now derived from the seven independent day values. */
export function getWeeklyTargetMinutes(dailyTargetMinutes: Record<DayKey, number>): number {
  return DAY_KEYS.reduce((sum, day) => sum + (dailyTargetMinutes[day] ?? 0), 0);
}

export const DEFAULT_TIME_SETTINGS: TimeSettingsState = {
  weekStart: 'monday',
  roundingMinutes: 15,
  dailyTargetMinutes: {
    sun: 0,
    mon: 8 * 60,
    tue: 8 * 60,
    wed: 8 * 60,
    thu: 8 * 60,
    fri: 8 * 60,
    sat: 4 * 60,
  },
  overtimeThresholdMinutes: 45 * 60,
  trackAttendance: true,
};

/**
 * v0 stored a single flat weeklyTargetMinutes. Preserve a workspace's choice
 * by spreading it evenly across Mon-Fri (the old implicit assumption in
 * AddHours' own dayTargetMinutes math) rather than silently reverting a
 * customized value back to the shipped default.
 */
export function migrateTimeSettingsV0(persisted: object): object {
  const old = persisted as { weeklyTargetMinutes?: number };
  const perWeekday = Math.max(1, Math.round((old.weeklyTargetMinutes ?? 40 * 60) / 5));
  return {
    ...persisted,
    dailyTargetMinutes: {
      sun: 0,
      mon: perWeekday,
      tue: perWeekday,
      wed: perWeekday,
      thu: perWeekday,
      fri: perWeekday,
      sat: 0,
    },
  };
}

/**
 * v1 stored ratesByUserId: a per-member hourly rate typed into Flow's settings
 * screen. Nothing ever read it back except the input that wrote it, so it bought
 * nothing, and HR already owns the rate (hr.employee.v1's hourly_rate). See
 * tryrtylr/rtylr#867.
 *
 * This deletes the key from the persisted blob rather than just dropping it from
 * the type. Leaving it would strand colleagues' pay rates in every existing
 * browser's localStorage indefinitely, which is the part of this worth
 * hurrying: the value was sensitive, unencrypted and useless at the same time.
 */
export function migrateTimeSettingsV1(persisted: object): object {
  const { ratesByUserId: _dropped, ...rest } = persisted as Record<string, unknown>;

  return rest;
}

export const useTimeSettingsStore = create<TimeSettingsState & TimeSettingsActions>()(
  persist(
    (set, get) => ({
      ...DEFAULT_TIME_SETTINGS,

      setWeekStart: (weekStart) => set({ weekStart }),
      setRoundingMinutes: (roundingMinutes) => set({ roundingMinutes }),
      setDailyTargetMinutes: (day, minutes) =>
        set({ dailyTargetMinutes: { ...get().dailyTargetMinutes, [day]: minutes } }),
      setOvertimeThresholdMinutes: (overtimeThresholdMinutes) => set({ overtimeThresholdMinutes }),
      setTrackAttendance: (trackAttendance) => set({ trackAttendance }),
    }),
    {
      name: TIME_SETTINGS_STORAGE_KEY,
      version: 3,
      // Keep working preferences from earlier versions and discard retired fields.
      migrate: (persisted, version) => {
        const afterV0 =
          version >= 1 ? (persisted as object) : migrateTimeSettingsV0(persisted as object);

        const afterV1 = version >= 2 ? afterV0 : migrateTimeSettingsV1(afterV0);
        const {
          approvalLocksEntry: _approval,
          autoLockDays: _autoLock,
          categories: _categories,
          ...preferences
        } = afterV1 as Record<string, unknown>;
        return preferences;
      },
    }
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
