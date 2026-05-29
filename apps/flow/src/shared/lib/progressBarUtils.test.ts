import { describe, it, expect } from 'vitest';

import {
  progressBarWidth,
  progressColor,
  healthColor,
  progressLabel,
  progressPercent,
  progressDescription,
} from './progressBarUtils';

describe('progressBarWidth', () => {
  it('returns percentage string', () => {
    expect(progressBarWidth(50)).toBe('50%');
  });

  it('clamps to 0-100', () => {
    expect(progressBarWidth(-10)).toBe('0%');
    expect(progressBarWidth(150)).toBe('100%');
  });
});

describe('progressColor', () => {
  it('returns success for 80+', () => {
    expect(progressColor(80)).toBe('success');
    expect(progressColor(100)).toBe('success');
  });

  it('returns warning for 50-79', () => {
    expect(progressColor(50)).toBe('warning');
    expect(progressColor(79)).toBe('warning');
  });

  it('returns danger for 1-49', () => {
    expect(progressColor(30)).toBe('danger');
  });

  it('returns default for 0', () => {
    expect(progressColor(0)).toBe('default');
  });
});

describe('healthColor', () => {
  it('returns green for on_track', () => {
    expect(healthColor('on_track')).toBe('green');
  });

  it('returns yellow for at_risk', () => {
    expect(healthColor('at_risk')).toBe('yellow');
  });

  it('returns red for off_track', () => {
    expect(healthColor('off_track')).toBe('red');
  });
});

describe('progressLabel', () => {
  it('returns completed / total', () => {
    expect(progressLabel(5, 10)).toBe('5 / 10');
  });

  it('returns 0 / 0 for empty', () => {
    expect(progressLabel(0, 0)).toBe('0 / 0');
  });
});

describe('progressPercent', () => {
  it('calculates percentage', () => {
    expect(progressPercent(3, 4)).toBe(75);
  });

  it('returns 0 for zero total', () => {
    expect(progressPercent(0, 0)).toBe(0);
  });
});

describe('progressDescription', () => {
  it('returns Not started for 0', () => {
    expect(progressDescription(0)).toBe('Not started');
  });

  it('returns Complete for 100', () => {
    expect(progressDescription(100)).toBe('Complete');
  });

  it('returns appropriate description for mid-range', () => {
    expect(progressDescription(50)).toBe('More than halfway');
    expect(progressDescription(90)).toBe('Almost done');
  });
});
