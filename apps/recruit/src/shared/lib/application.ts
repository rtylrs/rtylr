// Application tracking helpers for the recruit app admin.
// Pure functions that work on arrays of Submission objects and produce
// aggregated views used by the submissions dashboard.

import type { Submission } from './types';

export type StatusBucket = {
  status: string;
  count: number;
};

/**
 * countByStatus returns an array of {status, count} pairs representing
 * how many submissions exist per status. The result is sorted by count
 * descending.
 */
export function countByStatus(submissions: Submission[]): StatusBucket[] {
  const map = new Map<string, number>();
  for (const s of submissions) {
    map.set(s.status, (map.get(s.status) ?? 0) + 1);
  }
  return Array.from(map.entries())
    .map(([status, count]) => ({ status, count }))
    .sort((a, b) => b.count - a.count);
}

/**
 * countByCareer returns a map of careerId to submission count.
 * Useful for showing per-role applicant volume in the admin table.
 */
export function countByCareer(submissions: Submission[]): Map<number, number> {
  const map = new Map<number, number>();
  for (const s of submissions) {
    map.set(s.careerId, (map.get(s.careerId) ?? 0) + 1);
  }
  return map;
}

/**
 * latestSubmissions returns the n most recently created submissions.
 */
export function latestSubmissions(submissions: Submission[], n: number): Submission[] {
  return [...submissions]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, n);
}

/**
 * filterByStatus returns submissions matching the given status string.
 * Pass null or undefined to return all submissions.
 */
export function filterByStatus(
  submissions: Submission[],
  status: string | null | undefined
): Submission[] {
  if (!status) return submissions;
  return submissions.filter((s) => s.status === status);
}

/**
 * searchSubmissions performs a case-insensitive search across candidate
 * name, email, and career title.
 */
export function searchSubmissions(submissions: Submission[], query: string): Submission[] {
  const q = query.trim().toLowerCase();
  if (!q) return submissions;
  return submissions.filter(
    (s) =>
      s.candidateName.toLowerCase().includes(q) ||
      s.email.toLowerCase().includes(q) ||
      s.careerTitle.toLowerCase().includes(q)
  );
}

/**
 * conversionRate returns the percentage of submissions that reached
 * a terminal status (hired or rejected) relative to the total count.
 * Returns 0 when the submission list is empty.
 */
export function conversionRate(submissions: Submission[]): number {
  if (submissions.length === 0) return 0;
  const terminal = submissions.filter(
    (s) => s.status === 'hired' || s.status === 'rejected'
  ).length;
  return (terminal / submissions.length) * 100;
}

/**
 * hireRate returns the percentage of submissions that were hired.
 * Returns 0 for an empty list.
 */
export function hireRate(submissions: Submission[]): number {
  if (submissions.length === 0) return 0;
  const hired = submissions.filter((s) => s.status === 'hired').length;
  return (hired / submissions.length) * 100;
}
