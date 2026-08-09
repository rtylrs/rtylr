import { describe, it, expect } from 'vitest';
import {
  utilizationPct,
  utilizationLabel,
  scheduledHoursLabel,
  availableCapacity,
  capacityWarning,
} from './utilization';
import type { UtilizationResponse } from '../hooks/queries/types';

function makeUtil(overrides: Partial<UtilizationResponse> = {}): UtilizationResponse {
  return { scheduledHours: 160, availableHours: 200, utilizationPct: 80, ...overrides };
}

describe('utilizationPct', () => {
  it('returns the utilization percentage', () => {
    expect(utilizationPct(makeUtil({ utilizationPct: 75 }))).toBe(75);
  });

  it('clamps to 0', () => {
    expect(utilizationPct(makeUtil({ utilizationPct: -5 }))).toBe(0);
  });

  it('clamps to 100', () => {
    expect(utilizationPct(makeUtil({ utilizationPct: 110 }))).toBe(100);
  });
});

describe('utilizationLabel', () => {
  it('returns low for < 50%', () => {
    expect(utilizationLabel(49)).toBe('low');
  });

  it('returns normal for 50-79%', () => {
    expect(utilizationLabel(70)).toBe('normal');
  });

  it('returns high for 80-100%', () => {
    expect(utilizationLabel(90)).toBe('high');
  });

  it('returns overloaded for > 100%', () => {
    expect(utilizationLabel(110)).toBe('overloaded');
  });
});

describe('scheduledHoursLabel', () => {
  it('formats scheduled hours', () => {
    expect(scheduledHoursLabel(makeUtil({ scheduledHours: 120 }))).toBe('120h scheduled');
  });
});

describe('availableCapacity', () => {
  it('returns unused hours', () => {
    const util = makeUtil({ availableHours: 100, utilizationPct: 60 });
    expect(availableCapacity(util)).toBe(40);
  });

  it('returns 0 when fully utilized', () => {
    const util = makeUtil({ availableHours: 100, utilizationPct: 100 });
    expect(availableCapacity(util)).toBe(0);
  });
});

describe('capacityWarning', () => {
  it('returns true when utilization >= 80%', () => {
    expect(capacityWarning(makeUtil({ utilizationPct: 80 }))).toBe(true);
    expect(capacityWarning(makeUtil({ utilizationPct: 95 }))).toBe(true);
  });

  it('returns false when utilization < 80%', () => {
    expect(capacityWarning(makeUtil({ utilizationPct: 79 }))).toBe(false);
  });
});
