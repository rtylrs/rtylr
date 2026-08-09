// Job board helpers for the recruit careers listing feature.

import type { RecruitCareer } from './types';

/**
 * publishedCareers returns only careers that are both published and active.
 */
export function publishedCareers(careers: RecruitCareer[]): RecruitCareer[] {
  return careers.filter((c) => c.isPublished && c.isActive);
}

/**
 * careersByDepartment groups careers by their department field.
 */
export function careersByDepartment(careers: RecruitCareer[]): Map<string, RecruitCareer[]> {
  const map = new Map<string, RecruitCareer[]>();
  for (const career of careers) {
    const existing = map.get(career.department) ?? [];
    map.set(career.department, [...existing, career]);
  }
  return map;
}

/**
 * careersByLocation groups careers by their location field.
 */
export function careersByLocation(careers: RecruitCareer[]): Map<string, RecruitCareer[]> {
  const map = new Map<string, RecruitCareer[]>();
  for (const career of careers) {
    const existing = map.get(career.location) ?? [];
    map.set(career.location, [...existing, career]);
  }
  return map;
}

/**
 * filterByWorkMode returns careers matching the given work mode.
 */
export function filterByWorkMode(careers: RecruitCareer[], workMode: string): RecruitCareer[] {
  return careers.filter((c) => c.workMode === workMode);
}

/**
 * filterByEmploymentType returns careers matching the given employment type.
 */
export function filterByEmploymentType(careers: RecruitCareer[], type: string): RecruitCareer[] {
  return careers.filter((c) => c.employmentType === type);
}

/**
 * uniqueDepartments returns a sorted list of unique department names.
 */
export function uniqueDepartments(careers: RecruitCareer[]): string[] {
  return [...new Set(careers.map((c) => c.department))].sort();
}
