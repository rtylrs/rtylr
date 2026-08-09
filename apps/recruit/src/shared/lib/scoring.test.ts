import { describe, it, expect } from 'vitest';
import {
  completenessScore,
  completenessReasons,
  scoreSubmission,
  rankSubmissions,
} from './scoring';
import type { Submission } from './types';

function makeSubmission(overrides: Partial<Submission> = {}): Submission {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    candidateName: 'Alice',
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

describe('completenessScore', () => {
  it('returns 0 for empty submission', () => {
    expect(completenessScore(makeSubmission())).toBe(0);
  });

  it('returns 100 for fully filled submission', () => {
    const s = makeSubmission({
      resumeUrl: 'https://r.pdf',
      coverLetter: 'Hi',
      phone: '555',
    });
    expect(completenessScore(s)).toBe(100);
  });

  it('returns partial score', () => {
    const s = makeSubmission({ resumeUrl: 'https://r.pdf' });
    expect(completenessScore(s)).toBeGreaterThan(0);
    expect(completenessScore(s)).toBeLessThan(100);
  });
});

describe('completenessReasons', () => {
  it('lists present fields', () => {
    const s = makeSubmission({ resumeUrl: 'https://r.pdf', coverLetter: 'Hi' });
    const reasons = completenessReasons(s);
    expect(reasons).toContain('Resume attached');
    expect(reasons).toContain('Cover letter provided');
    expect(reasons).not.toContain('Phone number provided');
  });

  it('returns empty array for empty submission', () => {
    expect(completenessReasons(makeSubmission())).toHaveLength(0);
  });
});

describe('scoreSubmission', () => {
  it('returns a scored submission with reasons', () => {
    const s = makeSubmission({ resumeUrl: 'https://r.pdf' });
    const scored = scoreSubmission(s);
    expect(scored.submission).toBe(s);
    expect(scored.score).toBeGreaterThan(0);
    expect(scored.reasons).toContain('Resume attached');
  });
});

describe('rankSubmissions', () => {
  it('sorts by score descending', () => {
    const low = makeSubmission({ id: 1 });
    const high = makeSubmission({ id: 2, resumeUrl: 'r.pdf', coverLetter: 'Hi', phone: '555' });
    expect(rankSubmissions([low, high])[0].submission.id).toBe(2);
  });
});
