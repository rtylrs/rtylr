import { describe, expect, it } from 'vitest';
import { bucketCashFlow } from './bucketCashFlow';

const days = (from: string, count: number) => {
  const [y, m, d] = from.split('-').map(Number);
  return Array.from({ length: count }, (_, i) => {
    const date = new Date(y, m - 1, d + i);
    const iso = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
    return { month: iso, inflow: 1, outflow: 2 };
  });
};

describe('bucketCashFlow', () => {
  it('keeps a month daily with a readable sparse axis', () => {
    const buckets = bucketCashFlow(days('2026-10-01', 31), 'month');
    expect(buckets).toHaveLength(31);
    expect(buckets.filter((b) => b.label !== '').map((b) => b.label)).toEqual([
      '1',
      '5',
      '10',
      '15',
      '20',
      '25',
      '30',
    ]);
  });

  it('groups a year into twelve months', () => {
    const buckets = bucketCashFlow(days('2026-01-01', 365), 'year');
    expect(buckets).toHaveLength(12);
    expect(buckets[0]).toMatchObject({ label: 'Jan', inflow: 31, outflow: 62 });
    expect(buckets[11].label).toBe('Dec');
  });

  it('groups a quarter into Monday-start weeks', () => {
    const buckets = bucketCashFlow(days('2026-10-01', 92), 'quarter');
    expect(buckets[0].label).toBe('Sep 28');
    expect(buckets[1].label).toBe('Oct 5');
    expect(buckets.reduce((s, b) => s + b.inflow, 0)).toBe(92);
  });
});
