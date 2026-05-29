import { describe, it, expect } from 'vitest';
import {
  totalSessions,
  deviceSessionShare,
  topDeviceByRevenue,
  revenuePerSession,
  sortDevicesByRevenue,
  mobileShare,
} from './device';
import type { DeviceBreakdown } from './device';

function makeDevice(type: 'desktop' | 'mobile' | 'tablet', sessions: number, revenue: number): DeviceBreakdown {
  return { type, sessions, revenue };
}

const devices = [
  makeDevice('desktop', 600, 9000),
  makeDevice('mobile', 300, 2000),
  makeDevice('tablet', 100, 1000),
];

describe('totalSessions', () => {
  it('sums sessions across all devices', () => {
    expect(totalSessions(devices)).toBe(1000);
  });

  it('returns 0 for empty array', () => {
    expect(totalSessions([])).toBe(0);
  });
});

describe('deviceSessionShare', () => {
  it('returns percentage share of total sessions', () => {
    expect(deviceSessionShare(devices[0], devices)).toBe(60);
  });

  it('returns 0 when total sessions is 0', () => {
    const zero = [makeDevice('desktop', 0, 0)];
    expect(deviceSessionShare(zero[0], zero)).toBe(0);
  });
});

describe('topDeviceByRevenue', () => {
  it('returns device with highest revenue', () => {
    expect(topDeviceByRevenue(devices)?.type).toBe('desktop');
  });

  it('returns undefined for empty array', () => {
    expect(topDeviceByRevenue([])).toBeUndefined();
  });
});

describe('revenuePerSession', () => {
  it('calculates revenue per session', () => {
    expect(revenuePerSession(devices[0])).toBe(15);
  });

  it('returns 0 when sessions is 0', () => {
    expect(revenuePerSession(makeDevice('desktop', 0, 500))).toBe(0);
  });
});

describe('sortDevicesByRevenue', () => {
  it('sorts descending by revenue', () => {
    const sorted = sortDevicesByRevenue(devices);
    expect(sorted[0].type).toBe('desktop');
    expect(sorted[2].type).toBe('tablet');
  });

  it('does not mutate original', () => {
    sortDevicesByRevenue(devices);
    expect(devices[0].type).toBe('desktop');
  });
});

describe('mobileShare', () => {
  it('returns combined share of mobile and tablet', () => {
    expect(mobileShare(devices)).toBe(40);
  });

  it('returns 0 for empty array', () => {
    expect(mobileShare([])).toBe(0);
  });
});
