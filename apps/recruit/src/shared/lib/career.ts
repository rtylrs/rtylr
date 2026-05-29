// Career helpers for the recruit app.
// Pure functions that operate on RecruitCareer objects from types.ts.

import type { RecruitCareer } from './types';

// Canonical seniority ordering from junior to executive.
const SENIORITY_ORDER = ['intern', 'junior', 'mid', 'senior', 'lead', 'principal', 'executive'];

// Canonical employment type labels.
const EMPLOYMENT_LABELS: Record<string, string> = {
  'full-time': 'Full-time',
  'part-time': 'Part-time',
  contract: 'Contract',
  freelance: 'Freelance',
  internship: 'Internship',
};

// Canonical work mode labels.
const WORK_MODE_LABELS: Record<string, string> = {
  remote: 'Remote',
  hybrid: 'Hybrid',
  onsite: 'On-site',
  'on-site': 'On-site',
};

/**
 * Returns a human-readable label for an employment type string.
 * Falls back to capitalizing the raw value.
 */
export function employmentTypeLabel(type: string): string {
  return EMPLOYMENT_LABELS[type.toLowerCase()] ?? capitalize(type);
}

/**
 * Returns a human-readable label for a work mode string.
 */
export function workModeLabel(mode: string): string {
  return WORK_MODE_LABELS[mode.toLowerCase()] ?? capitalize(mode);
}

/**
 * Capitalizes the first letter of a string.
 */
function capitalize(s: string): string {
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/**
 * Formats a salary range as a human-readable string.
 * Returns an empty string when both min and max are 0.
 */
export function formatSalaryRange(
  min: number,
  max: number,
  currency = 'USD',
  locale = 'en-US'
): string {
  if (!min && !max) return '';
  const fmt = new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  });
  if (!max) return `${fmt.format(min)}+`;
  if (!min) return `Up to ${fmt.format(max)}`;
  return `${fmt.format(min)} - ${fmt.format(max)}`;
}

/**
 * Returns true when the career is both published and active (visible to candidates).
 */
export function isCareerVisible(career: RecruitCareer): boolean {
  return career.isPublished && career.isActive;
}

/**
 * Returns the numeric index of a seniority level (lower = more junior).
 * Unknown levels default to -1.
 */
export function seniorityIndex(seniority: string): number {
  return SENIORITY_ORDER.indexOf(seniority.toLowerCase());
}

/**
 * Sorts careers by seniority (senior levels first), then alphabetically by title.
 */
export function sortCareersBySeniority(careers: RecruitCareer[]): RecruitCareer[] {
  return [...careers].sort((a, b) => {
    const ai = seniorityIndex(a.seniority);
    const bi = seniorityIndex(b.seniority);
    const aIdx = ai === -1 ? SENIORITY_ORDER.length : ai;
    const bIdx = bi === -1 ? SENIORITY_ORDER.length : bi;
    if (aIdx !== bIdx) return bIdx - aIdx;
    return a.title.localeCompare(b.title);
  });
}

/**
 * Groups careers by department. Returns a Map of department name to career list.
 */
export function groupCareersByDepartment(careers: RecruitCareer[]): Map<string, RecruitCareer[]> {
  const map = new Map<string, RecruitCareer[]>();
  for (const career of careers) {
    const dept = career.department || 'Other';
    if (!map.has(dept)) {
      map.set(dept, []);
    }
    map.get(dept)!.push(career);
  }
  return map;
}

/**
 * Filters careers by visibility (published and active).
 */
export function filterVisibleCareers(careers: RecruitCareer[]): RecruitCareer[] {
  return careers.filter(isCareerVisible);
}

/**
 * Returns unique department names from a list of careers, sorted alphabetically.
 */
export function uniqueDepartments(careers: RecruitCareer[]): string[] {
  const seen = new Set<string>();
  for (const c of careers) {
    if (c.department) seen.add(c.department);
  }
  return [...seen].sort();
}
