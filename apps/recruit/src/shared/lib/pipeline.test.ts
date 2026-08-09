import { describe, it, expect } from 'vitest';
import {
  buildPipeline,
  pipelineConversionRates,
  overallConversionRate,
  avgTimeToStatus,
} from './pipeline';
import type { Submission } from './types';

function makeSubmission(overrides: Partial<Submission> = {}): Submission {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    candidateName: 'Test User',
    email: 'test@example.com',
    phone: '',
    resumeUrl: '',
    coverLetter: '',
    status: 'pending',
    answersJson: '',
    careerTitle: 'Engineer',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
    ...overrides,
  };
}

describe('buildPipeline', () => {
  it('returns stages in pipeline order', () => {
    const subs = [
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'reviewing' }),
      makeSubmission({ status: 'hired' }),
    ];
    const pipeline = buildPipeline(subs);
    expect(pipeline.map((s) => s.status)).toEqual(['pending', 'reviewing', 'shortlisted', 'hired']);
  });

  it('counts submissions per stage', () => {
    const subs = [
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'reviewing' }),
    ];
    const pipeline = buildPipeline(subs);
    expect(pipeline.find((s) => s.status === 'pending')?.count).toBe(2);
    expect(pipeline.find((s) => s.status === 'reviewing')?.count).toBe(1);
  });

  it('excludes rejected submissions', () => {
    const subs = [makeSubmission({ status: 'pending' }), makeSubmission({ status: 'rejected' })];
    const pipeline = buildPipeline(subs);
    // rejected is not included in any stage count
    const total = pipeline.reduce((sum, s) => sum + s.count, 0);
    expect(total).toBe(1);
  });

  it('returns zero counts for stages with no submissions', () => {
    const pipeline = buildPipeline([]);
    expect(pipeline.every((s) => s.count === 0)).toBe(true);
  });
});

describe('pipelineConversionRates', () => {
  it('calculates conversion from stage to stage', () => {
    const stages = [
      { status: 'pending', label: 'Applied', count: 100 },
      { status: 'reviewing', label: 'In Review', count: 50 },
      { status: 'shortlisted', label: 'Shortlisted', count: 20 },
      { status: 'hired', label: 'Hired', count: 5 },
    ];
    const rates = pipelineConversionRates(stages);
    expect(rates[0]).toBeCloseTo(50);
    expect(rates[1]).toBeCloseTo(40);
    expect(rates[2]).toBeCloseTo(25);
  });

  it('returns 0 for stages with zero count', () => {
    const stages = [
      { status: 'pending', label: 'Applied', count: 0 },
      { status: 'reviewing', label: 'In Review', count: 10 },
    ];
    expect(pipelineConversionRates(stages)[0]).toBe(0);
  });

  it('returns one fewer rate than stages', () => {
    const stages = [
      { status: 'a', label: 'A', count: 10 },
      { status: 'b', label: 'B', count: 5 },
      { status: 'c', label: 'C', count: 2 },
    ];
    expect(pipelineConversionRates(stages)).toHaveLength(2);
  });
});

describe('overallConversionRate', () => {
  it('calculates the hire rate as a percentage', () => {
    const subs = [
      makeSubmission({ status: 'hired' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
      makeSubmission({ status: 'pending' }),
    ];
    expect(overallConversionRate(subs)).toBeCloseTo(25);
  });

  it('returns 0 for empty input', () => {
    expect(overallConversionRate([])).toBe(0);
  });
});

describe('avgTimeToStatus', () => {
  it('calculates average days between creation and status update', () => {
    const subs = [
      makeSubmission({
        status: 'hired',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-03T00:00:00Z',
      }),
      makeSubmission({
        status: 'hired',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-05T00:00:00Z',
      }),
    ];
    expect(avgTimeToStatus(subs, 'hired')).toBeCloseTo(3);
  });

  it('returns 0 when no submissions match the target status', () => {
    const subs = [makeSubmission({ status: 'pending' })];
    expect(avgTimeToStatus(subs, 'hired')).toBe(0);
  });

  it('returns 0 for empty input', () => {
    expect(avgTimeToStatus([], 'hired')).toBe(0);
  });
});
