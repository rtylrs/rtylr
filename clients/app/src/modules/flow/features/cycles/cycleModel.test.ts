import { describe, expect, it } from 'vitest';
import { buildBurndown, endDateForLength } from './cycleModel';

describe('buildBurndown', () => {
  const items = [
    { weight: 1, completedAt: '2026-11-17T09:00:00' },
    { weight: 1, completedAt: '2026-11-18T23:30:00' },
    { weight: 1 },
    { weight: 1 },
  ];

  it('drops remaining work on the local day each task was finished', () => {
    const points = buildBurndown('2026-11-16', '2026-11-19', items, '2026-11-19');
    expect(points.map((p) => p.remaining)).toEqual([4, 3, 2, 2]);
    expect(points.map((p) => p.ideal)).toEqual([4, 2.7, 1.3, 0]);
  });

  it('leaves days after today without an actual value', () => {
    const points = buildBurndown('2026-11-16', '2026-11-19', items, '2026-11-17');
    expect(points.map((p) => p.remaining)).toEqual([4, 3, null, null]);
  });

  it('weighs items by their points', () => {
    const points = buildBurndown(
      '2026-11-16',
      '2026-11-17',
      [{ weight: 5, completedAt: '2026-11-16T10:00:00' }, { weight: 3 }],
      '2026-11-17'
    );
    expect(points.map((p) => p.remaining)).toEqual([3, 3]);
  });

  it('returns nothing for a cycle with no dates or inverted dates', () => {
    expect(buildBurndown('', '2026-11-19', items, '2026-11-19')).toEqual([]);
    expect(buildBurndown('2026-11-19', '2026-11-16', items, '2026-11-19')).toEqual([]);
  });
});

describe('endDateForLength', () => {
  it('counts the start day as the first day', () => {
    expect(endDateForLength('2026-11-16', 2)).toBe('2026-11-29');
    expect(endDateForLength('2026-12-28', 2)).toBe('2027-01-10');
  });
});
