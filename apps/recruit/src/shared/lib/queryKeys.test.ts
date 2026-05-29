import { describe, it, expect } from 'vitest';
import { recruitKeys } from './queryKeys';

describe('recruitKeys', () => {
  it('all starts with "recruit"', () => {
    expect(recruitKeys.all[0]).toBe('recruit');
  });

  it('publicCareers includes slug', () => {
    const key = recruitKeys.publicCareers('acme');
    expect(key).toContain('acme');
  });

  it('publicCareer includes both slugs', () => {
    const key = recruitKeys.publicCareer('acme', 'senior-engineer');
    expect(key).toContain('acme');
    expect(key).toContain('senior-engineer');
  });

  it('adminCareer includes the id', () => {
    const key = recruitKeys.adminCareer(42);
    expect(key).toContain(42);
  });

  it('adminSubmissions includes careerId in params', () => {
    const key = recruitKeys.adminSubmissions(5);
    const last = key[key.length - 1] as { careerId: number };
    expect(last.careerId).toBe(5);
  });

  it('adminFormFields includes careerId', () => {
    const key = recruitKeys.adminFormFields(7);
    expect(key).toContain(7);
  });

  it('public and admin keys are different', () => {
    expect(recruitKeys.public()).not.toEqual(recruitKeys.admin());
  });

  it('different admin resources produce different keys', () => {
    expect(recruitKeys.adminCareers()).not.toEqual(recruitKeys.adminSubmissions());
    expect(recruitKeys.adminCustomization()).not.toEqual(recruitKeys.adminSections());
  });
});
