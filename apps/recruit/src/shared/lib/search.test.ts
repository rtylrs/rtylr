import { describe, it, expect } from 'vitest';
import { searchCareers, searchSubmissions, highlightQuery, tokenizeQuery } from './search';
import type { RecruitCareer, Submission } from './types';

function makeCareer(overrides: Partial<RecruitCareer> = {}): RecruitCareer {
  return {
    id: 1,
    ownerId: 1,
    slug: 'engineer',
    title: 'Software Engineer',
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

function makeSubmission(overrides: Partial<Submission> = {}): Submission {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    candidateName: 'Alice Smith',
    email: 'alice@example.com',
    phone: '',
    resumeUrl: '',
    coverLetter: '',
    status: 'new',
    answersJson: '{}',
    careerTitle: 'Engineer',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('searchCareers', () => {
  it('filters by title', () => {
    const careers = [
      makeCareer({ title: 'Software Engineer' }),
      makeCareer({ title: 'Designer', department: 'Design' }),
    ];
    expect(searchCareers(careers, 'engineer')).toHaveLength(1);
  });

  it('filters by department', () => {
    const careers = [
      makeCareer({ department: 'Engineering' }),
      makeCareer({ department: 'Marketing' }),
    ];
    expect(searchCareers(careers, 'marketing')).toHaveLength(1);
  });

  it('returns all when query is empty', () => {
    const careers = [makeCareer(), makeCareer()];
    expect(searchCareers(careers, '')).toHaveLength(2);
  });

  it('is case-insensitive', () => {
    const careers = [makeCareer({ title: 'Designer', department: 'Design' })];
    expect(searchCareers(careers, 'DESIGNER')).toHaveLength(1);
  });
});

describe('searchSubmissions', () => {
  it('filters by candidate name', () => {
    const subs = [
      makeSubmission({ candidateName: 'Alice' }),
      makeSubmission({ candidateName: 'Bob', email: 'bob@example.com' }),
    ];
    expect(searchSubmissions(subs, 'alice')).toHaveLength(1);
  });

  it('filters by email', () => {
    const subs = [makeSubmission({ email: 'alice@example.com' })];
    expect(searchSubmissions(subs, 'alice@')).toHaveLength(1);
  });

  it('returns all for empty query', () => {
    expect(searchSubmissions([makeSubmission()], '')).toHaveLength(1);
  });
});

describe('highlightQuery', () => {
  it('wraps matching text with highlight markers', () => {
    const result = highlightQuery('Hello World', 'World');
    expect(result).toBe('Hello [HIGHLIGHT]World[/HIGHLIGHT]');
  });

  it('returns original text when no match', () => {
    expect(highlightQuery('Hello', 'xyz')).toBe('Hello');
  });

  it('returns original text for empty query', () => {
    expect(highlightQuery('Hello', '')).toBe('Hello');
  });
});

describe('tokenizeQuery', () => {
  it('splits by whitespace and lowercases', () => {
    expect(tokenizeQuery('Senior Engineer NYC')).toEqual(['senior', 'engineer', 'nyc']);
  });

  it('returns empty array for empty string', () => {
    expect(tokenizeQuery('')).toHaveLength(0);
  });

  it('removes extra whitespace', () => {
    expect(tokenizeQuery('  hello   world  ')).toEqual(['hello', 'world']);
  });
});
