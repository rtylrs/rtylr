import { describe, it, expect } from 'vitest';
import { validateCareerDraft, draftCompletionPct, salaryRangeLabel } from './careerDraft';

describe('validateCareerDraft', () => {
  it('returns isValid true for complete draft', () => {
    const result = validateCareerDraft({
      title: 'Engineer',
      department: 'Engineering',
      location: 'Remote',
      details: 'We are looking for...',
    });
    expect(result.isValid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('returns errors for missing required fields', () => {
    const result = validateCareerDraft({});
    expect(result.isValid).toBe(false);
    expect(result.errors).toContain('Title is required');
    expect(result.errors).toContain('Department is required');
  });

  it('errors when salaryMin exceeds salaryMax', () => {
    const result = validateCareerDraft({
      title: 'Engineer',
      department: 'Eng',
      location: 'Remote',
      details: 'Details',
      salaryMin: 100000,
      salaryMax: 80000,
    });
    expect(result.errors).toContain('Minimum salary cannot exceed maximum salary');
  });
});

describe('draftCompletionPct', () => {
  it('returns 0 for empty draft', () => {
    expect(draftCompletionPct({})).toBe(0);
  });

  it('returns 100 when all fields are filled', () => {
    expect(
      draftCompletionPct({
        title: 'Engineer',
        department: 'Eng',
        location: 'Remote',
        details: 'Details',
        requirements: 'Requirements',
        salaryMin: 80000,
        salaryMax: 120000,
      })
    ).toBe(100);
  });

  it('returns partial completion', () => {
    const pct = draftCompletionPct({ title: 'Engineer', department: 'Eng' });
    expect(pct).toBeGreaterThan(0);
    expect(pct).toBeLessThan(100);
  });
});

describe('salaryRangeLabel', () => {
  it('returns Competitive when no salary given', () => {
    expect(salaryRangeLabel(undefined, undefined)).toBe('Competitive');
  });

  it('returns range when both min and max given', () => {
    expect(salaryRangeLabel(80000, 120000)).toContain('80,000');
    expect(salaryRangeLabel(80000, 120000)).toContain('120,000');
  });

  it('returns plus format when only min given', () => {
    expect(salaryRangeLabel(80000, undefined)).toContain('80,000+');
  });

  it('returns up-to format when only max given', () => {
    expect(salaryRangeLabel(undefined, 120000)).toContain('120,000');
    expect(salaryRangeLabel(undefined, 120000)).toContain('Up to');
  });
});
