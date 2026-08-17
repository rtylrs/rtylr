/**
 * The day strip's own vocabulary for WHERE YOU WORKED (MyDay.tsx) and the
 * per-entry Location field (AddHours.tsx) must read identically - a hybrid
 * day is the whole reason the entry-level field exists, so "Remote" on one
 * and "Online" on the other would read as two different features rather than
 * one value at two levels. Shared here so both stay in sync by construction
 * instead of by two authors remembering to match wording.
 */
export const WORK_MODES = ['onsite', 'remote', 'hybrid'] as const;

export type WorkMode = (typeof WORK_MODES)[number];
