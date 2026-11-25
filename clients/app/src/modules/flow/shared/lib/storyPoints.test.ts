import { describe, expect, it } from 'vitest';
import { formatPoints, pointsOptions, toPointsScale } from './storyPoints';

describe('story points scales', () => {
  it('reads an unknown or missing scale as Fibonacci', () => {
    expect(toPointsScale(undefined)).toBe('fibonacci');
    expect(toPointsScale('hours')).toBe('fibonacci');
    expect(pointsOptions(toPointsScale('linear'))).toHaveLength(10);
  });

  it('shows T-shirt sizes for their stored numbers and falls back to the number', () => {
    expect(pointsOptions('tshirt').map((p) => formatPoints(p, 'tshirt'))).toEqual([
      'XS',
      'S',
      'M',
      'L',
      'XL',
    ]);
    // A value set on another scale before the project switched to T-shirt.
    expect(formatPoints(13, 'tshirt')).toBe('13');
    expect(formatPoints(13, 'fibonacci')).toBe('13');
  });
});
