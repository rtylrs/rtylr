import { describe, it, expect } from 'vitest';
import {
  sourceLabel,
  hireRate,
  topSourceByHires,
  fastestHiringSource,
  totalApplications,
  allCandidateSources,
} from './source';
import type { SourceMetric } from './source';

function makeMetric(
  source: 'job_board' | 'referral' | 'linkedin' | 'agency' | 'direct' | 'campus' | 'other',
  applications: number,
  hired: number,
  avgDaysToHire: number
): SourceMetric {
  return { source, applications, hired, avgDaysToHire };
}

const metrics = [
  makeMetric('referral', 50, 20, 15),
  makeMetric('job_board', 200, 30, 30),
  makeMetric('linkedin', 100, 10, 25),
];

describe('sourceLabel', () => {
  it('returns display label for known sources', () => {
    expect(sourceLabel('referral')).toBe('Referral');
    expect(sourceLabel('job_board')).toBe('Job Board');
  });

  it('returns raw value for unknown sources', () => {
    expect(sourceLabel('unknown')).toBe('unknown');
  });
});

describe('hireRate', () => {
  it('calculates hire rate percentage', () => {
    expect(hireRate(metrics[0])).toBe(40);
  });

  it('returns 0 when applications is 0', () => {
    expect(hireRate(makeMetric('direct', 0, 0, 0))).toBe(0);
  });
});

describe('topSourceByHires', () => {
  it('returns source with most hires', () => {
    expect(topSourceByHires(metrics)?.source).toBe('job_board');
  });

  it('returns undefined for empty array', () => {
    expect(topSourceByHires([])).toBeUndefined();
  });
});

describe('fastestHiringSource', () => {
  it('returns source with lowest avg days to hire', () => {
    expect(fastestHiringSource(metrics)?.source).toBe('referral');
  });

  it('ignores sources with 0 hires', () => {
    const noHires = [makeMetric('direct', 10, 0, 5), makeMetric('campus', 20, 5, 10)];
    expect(fastestHiringSource(noHires)?.source).toBe('campus');
  });

  it('returns undefined for empty array', () => {
    expect(fastestHiringSource([])).toBeUndefined();
  });
});

describe('totalApplications', () => {
  it('sums applications across sources', () => {
    expect(totalApplications(metrics)).toBe(350);
  });
});

describe('allCandidateSources', () => {
  it('returns all defined source values', () => {
    const sources = allCandidateSources();
    expect(sources).toContain('referral');
    expect(sources).toContain('job_board');
    expect(sources).toHaveLength(7);
  });
});
