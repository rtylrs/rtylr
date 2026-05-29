import { describe, it, expect } from 'vitest';
import { segmentSchema, SEGMENT_COLORS, isValidSegmentColor, isTrendingUp } from '../segments';

const baseSegment = {
  id: 1,
  name: 'VIP Customers',
  description: 'Top spenders',
  customerCount: 120,
  color: '#3b82f6',
  rules: ['spend > 500'],
  trend: '+5%',
  trendUp: true,
};

describe('segmentSchema', () => {
  it('parses a valid segment', () => {
    const result = segmentSchema.parse(baseSegment);
    expect(result.id).toBe(1);
    expect(result.name).toBe('VIP Customers');
    expect(result.trendUp).toBe(true);
  });

  it('coerces string id to number', () => {
    const result = segmentSchema.parse({ ...baseSegment, id: '5' });
    expect(result.id).toBe(5);
  });

  it('defaults rules to empty array when missing', () => {
    const { rules, ...rest } = baseSegment;
    const result = segmentSchema.parse(rest);
    expect(result.rules).toEqual([]);
  });

  it('allows optional icon', () => {
    const result = segmentSchema.parse({ ...baseSegment, icon: 'star' });
    expect(result.icon).toBe('star');
  });
});

describe('SEGMENT_COLORS', () => {
  it('contains blue and purple', () => {
    expect(SEGMENT_COLORS).toContain('#3b82f6');
    expect(SEGMENT_COLORS).toContain('#8b5cf6');
  });

  it('has 6 color options', () => {
    expect(SEGMENT_COLORS).toHaveLength(6);
  });
});

describe('isValidSegmentColor', () => {
  it('returns true for colors in the palette', () => {
    expect(isValidSegmentColor('#3b82f6')).toBe(true);
    expect(isValidSegmentColor('#ef4444')).toBe(true);
  });

  it('returns false for colors not in the palette', () => {
    expect(isValidSegmentColor('#ffffff')).toBe(false);
    expect(isValidSegmentColor('')).toBe(false);
  });
});

describe('isTrendingUp', () => {
  it('returns true when trendUp is true', () => {
    const segment = segmentSchema.parse(baseSegment);
    expect(isTrendingUp(segment)).toBe(true);
  });

  it('returns false when trendUp is false', () => {
    const segment = segmentSchema.parse({ ...baseSegment, trendUp: false });
    expect(isTrendingUp(segment)).toBe(false);
  });
});
