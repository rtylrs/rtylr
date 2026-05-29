import { describe, it, expect } from 'vitest';
import {
  channelConversionRate,
  channelROI,
  costPerConversion,
  topChannelByRevenue,
  sortChannelsByRevenue,
  channelLabel,
} from './channel';
import type { ChannelMetric } from './channel';

function makeMetric(
  channel: 'organic' | 'paid' | 'email' | 'social' | 'referral' | 'direct',
  sessions: number,
  conversions: number,
  spend: number,
  revenue: number
): ChannelMetric {
  return { channel, sessions, conversions, spend, revenue };
}

const metrics = [
  makeMetric('organic', 1000, 50, 0, 5000),
  makeMetric('paid', 500, 30, 1000, 3000),
  makeMetric('email', 300, 40, 200, 4000),
];

describe('channelConversionRate', () => {
  it('calculates conversion percentage', () => {
    expect(channelConversionRate(metrics[0])).toBe(5);
  });

  it('returns 0 when sessions is 0', () => {
    expect(channelConversionRate(makeMetric('direct', 0, 0, 0, 0))).toBe(0);
  });
});

describe('channelROI', () => {
  it('calculates ROI percentage', () => {
    expect(channelROI(metrics[1])).toBe(200);
  });

  it('returns 0 when spend is 0', () => {
    expect(channelROI(metrics[0])).toBe(0);
  });
});

describe('costPerConversion', () => {
  it('calculates cost per conversion', () => {
    expect(costPerConversion(metrics[1])).toBeCloseTo(33.33, 1);
  });

  it('returns 0 when conversions is 0', () => {
    expect(costPerConversion(makeMetric('paid', 100, 0, 500, 0))).toBe(0);
  });
});

describe('topChannelByRevenue', () => {
  it('returns channel with highest revenue', () => {
    expect(topChannelByRevenue(metrics)?.channel).toBe('organic');
  });

  it('returns undefined for empty array', () => {
    expect(topChannelByRevenue([])).toBeUndefined();
  });
});

describe('sortChannelsByRevenue', () => {
  it('sorts descending by revenue', () => {
    const sorted = sortChannelsByRevenue(metrics);
    expect(sorted[0].channel).toBe('organic');
    expect(sorted[2].channel).toBe('paid');
  });

  it('does not mutate original', () => {
    sortChannelsByRevenue(metrics);
    expect(metrics[0].channel).toBe('organic');
  });
});

describe('channelLabel', () => {
  it('returns display label for known channels', () => {
    expect(channelLabel('organic')).toBe('Organic Search');
    expect(channelLabel('paid')).toBe('Paid Ads');
  });

  it('returns raw value for unknown channels', () => {
    expect(channelLabel('unknown')).toBe('unknown');
  });
});
