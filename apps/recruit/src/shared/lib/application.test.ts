import { describe, it, expect } from 'vitest';
import {
  countByStatus,
  countByCareer,
  latestSubmissions,
  filterByStatus,
  searchSubmissions,
  conversionRate,
  hireRate,
} from './application';
import type { Submission } from './types';

function makeSubmission(overrides: Partial<Submission> = {}): Submission {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    candidateName: 'Jane Doe',
    email: 'jane@example.com',
    phone: '',
    resumeUrl: '',
    coverLetter: '',
    status: 'pending',
    answersJson: '',
    careerTitle: 'Software Engineer',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('countByStatus', () => {
  it('counts submissions by status', () => {
    const subs = [
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'hired' }),
    ];
    const result = countByStatus(subs);
    const pending = result.find((b) => b.status === 'pending');
    expect(pending?.count).toBe(2);
    const hired = result.find((b) => b.status === 'hired');
    expect(hired?.count).toBe(1);
  });

  it('sorts by count descending', () => {
    const subs = [
      makeSubmission({ status: 'reviewing' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
    ];
    const result = countByStatus(subs);
    expect(result[0].count).toBeGreaterThanOrEqual(result[1].count);
  });

  it('returns empty array for empty input', () => {
    expect(countByStatus([])).toEqual([]);
  });
});

describe('countByCareer', () => {
  it('counts submissions per careerId', () => {
    const subs = [
      makeSubmission({ careerId: 1 }),
      makeSubmission({ careerId: 2 }),
      makeSubmission({ careerId: 1 }),
    ];
    const map = countByCareer(subs);
    expect(map.get(1)).toBe(2);
    expect(map.get(2)).toBe(1);
  });

  it('returns an empty map for empty input', () => {
    expect(countByCareer([])).toEqual(new Map());
  });
});

describe('latestSubmissions', () => {
  it('returns the n most recent submissions', () => {
    const subs = [
      makeSubmission({ id: 1, createdAt: '2026-01-01T00:00:00Z' }),
      makeSubmission({ id: 2, createdAt: '2026-03-01T00:00:00Z' }),
      makeSubmission({ id: 3, createdAt: '2026-02-01T00:00:00Z' }),
    ];
    const result = latestSubmissions(subs, 2);
    expect(result[0].id).toBe(2);
    expect(result[1].id).toBe(3);
  });

  it('does not mutate the original array', () => {
    const subs = [makeSubmission({ id: 1, createdAt: '2026-01-01T00:00:00Z' })];
    latestSubmissions(subs, 1);
    expect(subs[0].id).toBe(1);
  });

  it('returns all when n exceeds length', () => {
    const subs = [makeSubmission()];
    expect(latestSubmissions(subs, 10)).toHaveLength(1);
  });
});

describe('filterByStatus', () => {
  it('filters by status', () => {
    const subs = [
      makeSubmission({ id: 1, status: 'pending' }),
      makeSubmission({ id: 2, status: 'hired' }),
    ];
    const result = filterByStatus(subs, 'pending');
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
  });

  it('returns all when status is null', () => {
    const subs = [makeSubmission(), makeSubmission()];
    expect(filterByStatus(subs, null)).toHaveLength(2);
  });

  it('returns all when status is undefined', () => {
    const subs = [makeSubmission(), makeSubmission()];
    expect(filterByStatus(subs, undefined)).toHaveLength(2);
  });
});

describe('searchSubmissions', () => {
  it('matches by candidate name case-insensitively', () => {
    const subs = [
      makeSubmission({ candidateName: 'Alice Smith' }),
      makeSubmission({ candidateName: 'Bob Jones' }),
    ];
    expect(searchSubmissions(subs, 'alice')).toHaveLength(1);
  });

  it('matches by email', () => {
    const subs = [makeSubmission({ email: 'alice@acme.com' }), makeSubmission()];
    expect(searchSubmissions(subs, 'acme')).toHaveLength(1);
  });

  it('matches by career title', () => {
    const subs = [
      makeSubmission({ careerTitle: 'Frontend Engineer' }),
      makeSubmission({ careerTitle: 'Backend Engineer' }),
    ];
    expect(searchSubmissions(subs, 'frontend')).toHaveLength(1);
  });

  it('returns all submissions for empty query', () => {
    const subs = [makeSubmission(), makeSubmission()];
    expect(searchSubmissions(subs, '')).toHaveLength(2);
  });
});

describe('conversionRate', () => {
  it('returns the percentage of terminal submissions', () => {
    const subs = [
      makeSubmission({ status: 'hired' }),
      makeSubmission({ status: 'rejected' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
    ];
    expect(conversionRate(subs)).toBeCloseTo(50);
  });

  it('returns 0 for empty input', () => {
    expect(conversionRate([])).toBe(0);
  });
});

describe('hireRate', () => {
  it('returns percentage of hired submissions', () => {
    const subs = [
      makeSubmission({ status: 'hired' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
    ];
    expect(hireRate(subs)).toBeCloseTo(25);
  });

  it('returns 0 for empty input', () => {
    expect(hireRate([])).toBe(0);
  });
});
