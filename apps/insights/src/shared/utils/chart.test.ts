import { describe, it, expect } from 'vitest';
import { reindexByLabel, topN, aggregateByLabel, fillMissingDates, pctShare } from './chart';

describe('reindexByLabel', () => {
  it('zips labels and values into LabeledPoint objects', () => {
    const result = reindexByLabel(['a', 'b', 'c'], [1, 2, 3]);
    expect(result).toEqual([
      { label: 'a', value: 1 },
      { label: 'b', value: 2 },
      { label: 'c', value: 3 },
    ]);
  });

  it('truncates to the shorter array', () => {
    const result = reindexByLabel(['a', 'b'], [10, 20, 30]);
    expect(result).toHaveLength(2);
  });

  it('returns an empty array when either input is empty', () => {
    expect(reindexByLabel([], [1, 2])).toEqual([]);
    expect(reindexByLabel(['a'], [])).toEqual([]);
  });
});

describe('topN', () => {
  it('returns the top n items sorted by value descending', () => {
    const points = [
      { label: 'a', value: 10 },
      { label: 'b', value: 50 },
      { label: 'c', value: 30 },
    ];
    const result = topN(points, 2);
    expect(result[0].label).toBe('b');
    expect(result[1].label).toBe('c');
  });

  it('returns all items when n exceeds the array length', () => {
    const points = [{ label: 'x', value: 5 }];
    expect(topN(points, 10)).toHaveLength(1);
  });

  it('does not mutate the original array', () => {
    const points = [
      { label: 'a', value: 1 },
      { label: 'b', value: 2 },
    ];
    topN(points, 1);
    expect(points[0].label).toBe('a');
  });

  it('returns an empty array for empty input', () => {
    expect(topN([], 3)).toEqual([]);
  });
});

describe('aggregateByLabel', () => {
  it('sums values that share the same label', () => {
    const points = [
      { label: '2026-01-01', value: 100 },
      { label: '2026-01-01', value: 200 },
      { label: '2026-01-02', value: 50 },
    ];
    const result = aggregateByLabel(points);
    const jan1 = result.find((p) => p.label === '2026-01-01');
    expect(jan1?.value).toBe(300);
    const jan2 = result.find((p) => p.label === '2026-01-02');
    expect(jan2?.value).toBe(50);
  });

  it('preserves unique labels unchanged', () => {
    const points = [
      { label: 'a', value: 5 },
      { label: 'b', value: 10 },
    ];
    expect(aggregateByLabel(points)).toHaveLength(2);
  });

  it('returns an empty array for empty input', () => {
    expect(aggregateByLabel([])).toEqual([]);
  });
});

describe('fillMissingDates', () => {
  it('inserts zero-value entries for missing dates', () => {
    const points = [{ label: '2026-01-01', value: 100 }];
    const result = fillMissingDates(points, '2026-01-01', '2026-01-03');
    expect(result).toHaveLength(3);
    expect(result[1]).toEqual({ label: '2026-01-02', value: 0 });
    expect(result[2]).toEqual({ label: '2026-01-03', value: 0 });
  });

  it('preserves existing values', () => {
    const points = [{ label: '2026-03-05', value: 42 }];
    const result = fillMissingDates(points, '2026-03-05', '2026-03-05');
    expect(result[0].value).toBe(42);
  });

  it('returns a single entry when from equals to', () => {
    const result = fillMissingDates([], '2026-05-01', '2026-05-01');
    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({ label: '2026-05-01', value: 0 });
  });
});

describe('pctShare', () => {
  it('converts values to percentage of total', () => {
    const result = pctShare([25, 75]);
    expect(result[0]).toBeCloseTo(25);
    expect(result[1]).toBeCloseTo(75);
  });

  it('returns all zeros when total is zero', () => {
    expect(pctShare([0, 0, 0])).toEqual([0, 0, 0]);
  });

  it('sums to 100 for positive values', () => {
    const result = pctShare([10, 20, 30, 40]);
    const sum = result.reduce((a, b) => a + b, 0);
    expect(sum).toBeCloseTo(100);
  });

  it('returns an empty array for empty input', () => {
    expect(pctShare([])).toEqual([]);
  });
});
