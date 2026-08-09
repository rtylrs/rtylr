import { describe, it, expect } from 'vitest';
import {
  totalCount,
  groupShare,
  diversityIndex,
  topGroup,
  filterByCategory,
  categoryNames,
} from './diversity';
import type { DiversityMetric } from './diversity';

function makeMetric(category: string, group: string, count: number): DiversityMetric {
  return { category, group, count };
}

const genderMetrics = [
  makeMetric('gender', 'Male', 60),
  makeMetric('gender', 'Female', 35),
  makeMetric('gender', 'Non-binary', 5),
];

const metrics = [
  ...genderMetrics,
  makeMetric('ethnicity', 'White', 55),
  makeMetric('ethnicity', 'Black', 20),
  makeMetric('ethnicity', 'Asian', 25),
];

describe('totalCount', () => {
  it('sums counts across all metrics', () => {
    expect(totalCount(genderMetrics)).toBe(100);
  });
});

describe('groupShare', () => {
  it('returns percentage share', () => {
    expect(groupShare(genderMetrics[0], genderMetrics)).toBe(60);
  });

  it('returns 0 when total is 0', () => {
    const zero = [makeMetric('g', 'X', 0)];
    expect(groupShare(zero[0], zero)).toBe(0);
  });
});

describe('diversityIndex', () => {
  it('returns a value between 0 and 1', () => {
    const index = diversityIndex(genderMetrics);
    expect(index).toBeGreaterThan(0);
    expect(index).toBeLessThan(1);
  });

  it('returns 0 for single group', () => {
    expect(diversityIndex([makeMetric('g', 'X', 100)])).toBe(0);
  });

  it('returns 0 for empty array', () => {
    expect(diversityIndex([])).toBe(0);
  });
});

describe('topGroup', () => {
  it('returns the group with the highest count', () => {
    expect(topGroup(genderMetrics)?.group).toBe('Male');
  });

  it('returns undefined for empty array', () => {
    expect(topGroup([])).toBeUndefined();
  });
});

describe('filterByCategory', () => {
  it('filters metrics by category', () => {
    const gender = filterByCategory(metrics, 'gender');
    expect(gender).toHaveLength(3);
    expect(gender.every((m) => m.category === 'gender')).toBe(true);
  });
});

describe('categoryNames', () => {
  it('returns unique category names', () => {
    const names = categoryNames(metrics);
    expect(names).toContain('gender');
    expect(names).toContain('ethnicity');
    expect(names).toHaveLength(2);
  });
});
