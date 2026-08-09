import { describe, it, expect } from 'vitest';
import {
  publishedCareers,
  careersByDepartment,
  careersByLocation,
  filterByWorkMode,
  filterByEmploymentType,
  uniqueDepartments,
} from './jobBoard';
import type { RecruitCareer } from './types';

function makeCareer(overrides: Partial<RecruitCareer> = {}): RecruitCareer {
  return {
    id: 1,
    ownerId: 1,
    slug: 'engineer',
    title: 'Engineer',
    department: 'Engineering',
    location: 'Remote',
    employmentType: 'full_time',
    workMode: 'remote',
    seniority: 'mid',
    details: '',
    requirements: '',
    salaryMin: 80000,
    salaryMax: 120000,
    salaryCurrency: 'USD',
    isPublished: true,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('publishedCareers', () => {
  it('returns only published and active careers', () => {
    const careers = [
      makeCareer({ id: 1, isPublished: true, isActive: true }),
      makeCareer({ id: 2, isPublished: false, isActive: true }),
      makeCareer({ id: 3, isPublished: true, isActive: false }),
    ];
    expect(publishedCareers(careers)).toHaveLength(1);
    expect(publishedCareers(careers)[0].id).toBe(1);
  });
});

describe('careersByDepartment', () => {
  it('groups careers by department', () => {
    const careers = [
      makeCareer({ id: 1, department: 'Engineering' }),
      makeCareer({ id: 2, department: 'Engineering' }),
      makeCareer({ id: 3, department: 'Marketing' }),
    ];
    const groups = careersByDepartment(careers);
    expect(groups.get('Engineering')).toHaveLength(2);
    expect(groups.get('Marketing')).toHaveLength(1);
  });
});

describe('careersByLocation', () => {
  it('groups careers by location', () => {
    const careers = [
      makeCareer({ id: 1, location: 'Remote' }),
      makeCareer({ id: 2, location: 'NYC' }),
      makeCareer({ id: 3, location: 'Remote' }),
    ];
    const groups = careersByLocation(careers);
    expect(groups.get('Remote')).toHaveLength(2);
  });
});

describe('filterByWorkMode', () => {
  it('filters by work mode', () => {
    const careers = [makeCareer({ workMode: 'remote' }), makeCareer({ workMode: 'onsite' })];
    expect(filterByWorkMode(careers, 'remote')).toHaveLength(1);
  });
});

describe('filterByEmploymentType', () => {
  it('filters by employment type', () => {
    const careers = [
      makeCareer({ employmentType: 'full_time' }),
      makeCareer({ employmentType: 'part_time' }),
    ];
    expect(filterByEmploymentType(careers, 'part_time')).toHaveLength(1);
  });
});

describe('uniqueDepartments', () => {
  it('returns sorted unique departments', () => {
    const careers = [
      makeCareer({ department: 'Engineering' }),
      makeCareer({ department: 'Marketing' }),
      makeCareer({ department: 'Engineering' }),
    ];
    expect(uniqueDepartments(careers)).toEqual(['Engineering', 'Marketing']);
  });
});
