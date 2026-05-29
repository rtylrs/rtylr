import { describe, it, expect } from 'vitest';
import {
  percentOf,
  periodChange,
  formatChange,
  isImprovement,
  average,
  median,
  maxValue,
  minValue,
  sum,
  groupSum,
} from './analyticsUtils';

describe('percentOf', () => {
  it('returns the percentage of a part in a total', () => {
    expect(percentOf(25, 100)).toBe(25);
  });

  it('returns 0 when total is 0', () => {
    expect(percentOf(5, 0)).toBe(0);
  });

  it('rounds to one decimal place', () => {
    expect(percentOf(1, 3)).toBe(33.3);
  });
});

describe('periodChange', () => {
  it('returns percentage change from previous to current', () => {
    expect(periodChange(110, 100)).toBe(10);
  });

  it('returns negative for a decrease', () => {
    expect(periodChange(80, 100)).toBe(-20);
  });

  it('returns null when previous is 0', () => {
    expect(periodChange(50, 0)).toBeNull();
  });
});

describe('formatChange', () => {
  it('formats a positive change with a plus sign', () => {
    expect(formatChange(12.5)).toBe('+12.5%');
  });

  it('formats a negative change without a plus sign', () => {
    expect(formatChange(-3.0)).toBe('-3.0%');
  });

  it('returns "N/A" for null', () => {
    expect(formatChange(null)).toBe('N/A');
  });
});

describe('isImprovement', () => {
  it('returns true for positive change when higherIsBetter', () => {
    expect(isImprovement(5, true)).toBe(true);
  });

  it('returns false for negative change when higherIsBetter', () => {
    expect(isImprovement(-5, true)).toBe(false);
  });

  it('returns true for negative change when lower is better', () => {
    expect(isImprovement(-5, false)).toBe(true);
  });

  it('returns false for null change', () => {
    expect(isImprovement(null)).toBe(false);
  });
});

describe('average', () => {
  it('returns the average of a number array', () => {
    expect(average([1, 2, 3, 4, 5])).toBe(3);
  });

  it('returns 0 for an empty array', () => {
    expect(average([])).toBe(0);
  });
});

describe('median', () => {
  it('returns the median of an odd-length array', () => {
    expect(median([1, 3, 5])).toBe(3);
  });

  it('returns the average of the two middle values for even-length arrays', () => {
    expect(median([1, 2, 3, 4])).toBe(2.5);
  });

  it('returns 0 for an empty array', () => {
    expect(median([])).toBe(0);
  });
});

describe('maxValue', () => {
  it('returns the largest value', () => {
    expect(maxValue([3, 1, 4, 1, 5, 9])).toBe(9);
  });

  it('returns 0 for an empty array', () => {
    expect(maxValue([])).toBe(0);
  });
});

describe('minValue', () => {
  it('returns the smallest value', () => {
    expect(minValue([3, 1, 4, 1, 5, 9])).toBe(1);
  });

  it('returns 0 for an empty array', () => {
    expect(minValue([])).toBe(0);
  });
});

describe('sum', () => {
  it('returns the sum of all values', () => {
    expect(sum([1, 2, 3, 4, 5])).toBe(15);
  });

  it('returns 0 for an empty array', () => {
    expect(sum([])).toBe(0);
  });
});

describe('groupSum', () => {
  const items = [
    { category: 'A', amount: 10 },
    { category: 'B', amount: 20 },
    { category: 'A', amount: 5 },
  ];

  it('sums values by key', () => {
    const result = groupSum(items, 'category', 'amount');
    expect(result['A']).toBe(15);
    expect(result['B']).toBe(20);
  });

  it('returns an empty object for an empty array', () => {
    expect(groupSum([], 'category', 'amount')).toEqual({});
  });
});
