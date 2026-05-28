import { describe, it, expect } from 'vitest';
import {
  statusColor,
  varianceColor,
  stockLevelColor,
  kdsTimerColor,
  loyaltyTierColor,
} from './colorUtils';

describe('statusColor', () => {
  it('returns "green" for success', () => {
    expect(statusColor('success')).toBe('green');
  });

  it('returns "red" for error', () => {
    expect(statusColor('error')).toBe('red');
  });

  it('returns "yellow" for warning', () => {
    expect(statusColor('warning')).toBe('yellow');
  });

  it('returns "blue" for info', () => {
    expect(statusColor('info')).toBe('blue');
  });

  it('returns "gray" for neutral', () => {
    expect(statusColor('neutral')).toBe('gray');
  });
});

describe('varianceColor', () => {
  it('returns "green" for a positive variance', () => {
    expect(varianceColor(10)).toBe('green');
  });

  it('returns "red" for a negative variance', () => {
    expect(varianceColor(-5)).toBe('red');
  });

  it('returns "gray" for zero variance', () => {
    expect(varianceColor(0)).toBe('gray');
  });

  it('returns "gray" for NaN', () => {
    expect(varianceColor(NaN)).toBe('gray');
  });
});

describe('stockLevelColor', () => {
  it('returns "red" for zero stock', () => {
    expect(stockLevelColor(0, 5)).toBe('red');
  });

  it('returns "yellow" for low stock', () => {
    expect(stockLevelColor(3, 5)).toBe('yellow');
  });

  it('returns "green" for adequate stock', () => {
    expect(stockLevelColor(20, 5)).toBe('green');
  });
});

describe('kdsTimerColor', () => {
  it('returns "green" for under 5 minutes', () => {
    expect(kdsTimerColor(200)).toBe('green');
  });

  it('returns "yellow" for 5-10 minutes', () => {
    expect(kdsTimerColor(400)).toBe('yellow');
  });

  it('returns "red" for over 10 minutes', () => {
    expect(kdsTimerColor(700)).toBe('red');
  });
});

describe('loyaltyTierColor', () => {
  it('returns "gray" for fewer than 500 points', () => {
    expect(loyaltyTierColor(100)).toBe('gray');
  });

  it('returns "yellow" for 500-1999 points', () => {
    expect(loyaltyTierColor(1000)).toBe('yellow');
  });

  it('returns "orange" for 2000-4999 points', () => {
    expect(loyaltyTierColor(3000)).toBe('orange');
  });

  it('returns "purple" for 5000+ points', () => {
    expect(loyaltyTierColor(6000)).toBe('purple');
  });
});
