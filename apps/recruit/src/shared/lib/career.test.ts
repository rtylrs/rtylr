import { describe, it, expect } from 'vitest';
import {
  employmentTypeLabel,
  workModeLabel,
  formatSalaryRange,
  isCareerVisible,
  seniorityIndex,
  sortCareersBySeniority,
  groupCareersByDepartment,
  filterVisibleCareers,
  uniqueDepartments,
} from './career';
import type { RecruitCareer } from './types';

function makeCareer(overrides: Partial<RecruitCareer> = {}): RecruitCareer {
  return {
    id: 1,
    ownerId: 1,
    slug: 'test-role',
    title: 'Software Engineer',
    department: 'Engineering',
    location: 'Remote',
    employmentType: 'full-time',
    workMode: 'remote',
    seniority: 'senior',
    details: '',
    requirements: '',
    salaryMin: 0,
    salaryMax: 0,
    salaryCurrency: 'USD',
    isPublished: true,
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('employmentTypeLabel', () => {
  it('returns Full-time for full-time', () => {
    expect(employmentTypeLabel('full-time')).toBe('Full-time');
  });

  it('returns Part-time for part-time', () => {
    expect(employmentTypeLabel('part-time')).toBe('Part-time');
  });

  it('returns Contract for contract', () => {
    expect(employmentTypeLabel('contract')).toBe('Contract');
  });

  it('capitalizes unknown types', () => {
    expect(employmentTypeLabel('temporary')).toBe('Temporary');
  });
});

describe('workModeLabel', () => {
  it('returns Remote for remote', () => {
    expect(workModeLabel('remote')).toBe('Remote');
  });

  it('returns Hybrid for hybrid', () => {
    expect(workModeLabel('hybrid')).toBe('Hybrid');
  });

  it('returns On-site for onsite', () => {
    expect(workModeLabel('onsite')).toBe('On-site');
  });

  it('returns On-site for on-site', () => {
    expect(workModeLabel('on-site')).toBe('On-site');
  });

  it('capitalizes unknown modes', () => {
    expect(workModeLabel('flexible')).toBe('Flexible');
  });
});

describe('formatSalaryRange', () => {
  it('returns empty string when both min and max are 0', () => {
    expect(formatSalaryRange(0, 0)).toBe('');
  });

  it('formats a range with both min and max', () => {
    const result = formatSalaryRange(80000, 120000, 'USD');
    expect(result).toContain('$80,000');
    expect(result).toContain('$120,000');
    expect(result).toContain(' - ');
  });

  it('shows a plus sign when max is 0', () => {
    const result = formatSalaryRange(100000, 0, 'USD');
    expect(result).toContain('$100,000');
    expect(result).toContain('+');
  });

  it('shows "Up to" when min is 0', () => {
    const result = formatSalaryRange(0, 150000, 'USD');
    expect(result).toContain('Up to');
    expect(result).toContain('$150,000');
  });
});

describe('isCareerVisible', () => {
  it('returns true when both isPublished and isActive are true', () => {
    expect(isCareerVisible(makeCareer())).toBe(true);
  });

  it('returns false when isPublished is false', () => {
    expect(isCareerVisible(makeCareer({ isPublished: false }))).toBe(false);
  });

  it('returns false when isActive is false', () => {
    expect(isCareerVisible(makeCareer({ isActive: false }))).toBe(false);
  });
});

describe('seniorityIndex', () => {
  it('returns a higher index for senior than junior', () => {
    expect(seniorityIndex('senior')).toBeGreaterThan(seniorityIndex('junior'));
  });

  it('returns -1 for unknown seniority', () => {
    expect(seniorityIndex('wizard')).toBe(-1);
  });

  it('is case-insensitive', () => {
    expect(seniorityIndex('Senior')).toBe(seniorityIndex('senior'));
  });
});

describe('sortCareersBySeniority', () => {
  it('sorts senior before junior', () => {
    const careers = [
      makeCareer({ id: 1, seniority: 'junior', title: 'A' }),
      makeCareer({ id: 2, seniority: 'senior', title: 'B' }),
    ];
    const sorted = sortCareersBySeniority(careers);
    expect(sorted[0].id).toBe(2);
    expect(sorted[1].id).toBe(1);
  });

  it('sorts alphabetically by title within the same seniority', () => {
    const careers = [
      makeCareer({ id: 1, seniority: 'senior', title: 'Zebra Engineer' }),
      makeCareer({ id: 2, seniority: 'senior', title: 'Alpha Engineer' }),
    ];
    const sorted = sortCareersBySeniority(careers);
    expect(sorted[0].id).toBe(2);
    expect(sorted[1].id).toBe(1);
  });

  it('does not mutate the original array', () => {
    const careers = [makeCareer({ id: 1, seniority: 'junior' })];
    sortCareersBySeniority(careers);
    expect(careers[0].id).toBe(1);
  });
});

describe('groupCareersByDepartment', () => {
  it('groups careers by department', () => {
    const careers = [
      makeCareer({ id: 1, department: 'Engineering' }),
      makeCareer({ id: 2, department: 'Design' }),
      makeCareer({ id: 3, department: 'Engineering' }),
    ];
    const map = groupCareersByDepartment(careers);
    expect(map.get('Engineering')).toHaveLength(2);
    expect(map.get('Design')).toHaveLength(1);
  });

  it('uses "Other" for careers with empty department', () => {
    const careers = [makeCareer({ department: '' })];
    const map = groupCareersByDepartment(careers);
    expect(map.has('Other')).toBe(true);
  });
});

describe('filterVisibleCareers', () => {
  it('returns only published and active careers', () => {
    const careers = [
      makeCareer({ id: 1, isPublished: true, isActive: true }),
      makeCareer({ id: 2, isPublished: false, isActive: true }),
      makeCareer({ id: 3, isPublished: true, isActive: false }),
    ];
    const result = filterVisibleCareers(careers);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
  });
});

describe('uniqueDepartments', () => {
  it('returns unique departments sorted alphabetically', () => {
    const careers = [
      makeCareer({ department: 'Product' }),
      makeCareer({ department: 'Engineering' }),
      makeCareer({ department: 'Product' }),
    ];
    expect(uniqueDepartments(careers)).toEqual(['Engineering', 'Product']);
  });

  it('excludes empty department strings', () => {
    const careers = [makeCareer({ department: '' }), makeCareer({ department: 'Design' })];
    expect(uniqueDepartments(careers)).toEqual(['Design']);
  });
});
