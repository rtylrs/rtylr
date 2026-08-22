import { describe, it, expect } from 'vitest';
import { segmentSchema, SEGMENT_COLORS, isValidSegmentColor, isTrendingUp } from '../segments';

const baseSegment = {
  id: '1',
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
    expect(result.id).toBe('1');
    expect(result.name).toBe('VIP Customers');
    expect(result.trendUp).toBe(true);
  });

  it('keeps id as a string', () => {
    const result = segmentSchema.parse({ ...baseSegment, id: '5' });
    expect(result.id).toBe('5');
  });

  it('defaults rules to empty array when missing', () => {
    const { rules: _rules, ...rest } = baseSegment;
    const result = segmentSchema.parse(rest);
    expect(result.rules).toEqual([]);
  });

  it('allows optional icon', () => {
    const result = segmentSchema.parse({ ...baseSegment, icon: 'star' });
    expect(result.icon).toBe('star');
  });

  it('parses a backend wire payload and defaults presentation fields', () => {
    const result = segmentSchema.parse({
      id: '7',
      ownerId: '1',
      name: 'New Segment',
      description: 'From the API',
      filters: '{"tier":"gold"}',
      customerCount: '42',
      createdAt: '2024-06-01T00:00:00Z',
      updatedAt: '2024-06-02T00:00:00Z',
    });
    expect(result.id).toBe('7');
    expect(result.filters).toEqual('{"tier":"gold"}');
    expect(result.customerCount).toBe(42);
    expect(result.color).toBe('#3b82f6');
    expect(result.rules).toEqual([]);
    expect(result.trendUp).toBe(false);
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
