import type { Milestone, Project } from '../stores/flowStore';

export interface DateRange {
  start: string;
  end: string;
}

/** Returns true when two date ranges overlap. */
export function rangesOverlap(a: DateRange, b: DateRange): boolean {
  return a.start <= b.end && b.start <= a.end;
}

/** Clamp a date string to be within [min, max]. */
export function clampToRange(date: string, min: string, max: string): string {
  if (date < min) return min;
  if (date > max) return max;
  return date;
}

/** Returns the span in days between two ISO date strings. */
export function spanInDays(start: string, end: string): number {
  const diff = new Date(end).getTime() - new Date(start).getTime();
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)));
}

/** Returns milestones that fall within a given date range. */
export function milestonesInRange(milestones: Milestone[], range: DateRange): Milestone[] {
  return milestones.filter((m) => m.targetDate >= range.start && m.targetDate <= range.end);
}

/** Returns projects whose start/target date overlaps with the range. */
export function projectsInRange(projects: Project[], range: DateRange): Project[] {
  return projects.filter((p) => rangesOverlap({ start: p.startDate, end: p.targetDate }, range));
}

/** Returns the earliest start date among all projects (or empty string). */
export function earliestProjectStart(projects: Project[]): string {
  if (projects.length === 0) return '';
  return projects.reduce(
    (min, p) => (p.startDate < min ? p.startDate : min),
    projects[0].startDate
  );
}

/** Returns the latest target date among all projects (or empty string). */
export function latestProjectEnd(projects: Project[]): string {
  if (projects.length === 0) return '';
  return projects.reduce(
    (max, p) => (p.targetDate > max ? p.targetDate : max),
    projects[0].targetDate
  );
}
