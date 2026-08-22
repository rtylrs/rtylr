import { describe, expect, it } from 'vitest';
import { mergeSegmentData } from '../useSegments';
import type { Segment } from '../../../schemas/segments';

const existing: Segment = {
  id: '7',
  name: 'High value',
  description: 'Customers with high spend',
  filters: JSON.stringify({ rules: ['totalSpent > 500'], matchAll: true }),
  customerCount: 12,
  color: '#3b82f6',
  rules: ['totalSpent > 500'],
  trend: '+4%',
  trendUp: true,
  matchAll: true,
};

describe('mergeSegmentData', () => {
  it('preserves filters and presentation fields for partial updates', () => {
    expect(mergeSegmentData(existing, { name: 'VIP customers' })).toEqual({
      ...existing,
      name: 'VIP customers',
    });
  });
});
