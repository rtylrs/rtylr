import { describe, it, expect } from 'vitest';
import {
  countBadge,
  shouldShowBadge,
  truncateBadgeLabel,
  colorForString,
  scoreColor,
  makeBadge,
} from './badgeUtils';

describe('countBadge', () => {
  it('returns "0" for zero count', () => {
    expect(countBadge(0)).toBe('0');
  });

  it('returns the count as a string for small numbers', () => {
    expect(countBadge(5)).toBe('5');
  });

  it('caps at the default max of 99', () => {
    expect(countBadge(100)).toBe('99+');
  });

  it('uses a custom max', () => {
    expect(countBadge(10, 9)).toBe('9+');
  });

  it('does not cap when exactly at max', () => {
    expect(countBadge(99)).toBe('99');
  });
});

describe('shouldShowBadge', () => {
  it('returns true for count > 0', () => {
    expect(shouldShowBadge(1)).toBe(true);
  });

  it('returns false for count 0', () => {
    expect(shouldShowBadge(0)).toBe(false);
  });

  it('returns false for negative count', () => {
    expect(shouldShowBadge(-1)).toBe(false);
  });
});

describe('truncateBadgeLabel', () => {
  it('returns the label unchanged when within limit', () => {
    expect(truncateBadgeLabel('Short', 20)).toBe('Short');
  });

  it('truncates long labels and adds ellipsis', () => {
    const result = truncateBadgeLabel('This is a very long label', 10);
    expect(result).toContain('...');
    expect(result.startsWith('This is a')).toBe(true);
  });
});

describe('colorForString', () => {
  it('returns a valid badge color', () => {
    const validColors = ['blue', 'green', 'yellow', 'red', 'purple', 'gray', 'orange', 'teal'];
    expect(validColors).toContain(colorForString('hello'));
  });

  it('returns the same color for the same string', () => {
    expect(colorForString('test')).toBe(colorForString('test'));
  });
});

describe('scoreColor', () => {
  it('returns green for score >= 80', () => {
    expect(scoreColor(100)).toBe('green');
    expect(scoreColor(80)).toBe('green');
  });

  it('returns teal for score 60-79', () => {
    expect(scoreColor(60)).toBe('teal');
  });

  it('returns yellow for score 40-59', () => {
    expect(scoreColor(40)).toBe('yellow');
  });

  it('returns orange for score 20-39', () => {
    expect(scoreColor(20)).toBe('orange');
  });

  it('returns red for score < 20', () => {
    expect(scoreColor(0)).toBe('red');
  });
});

describe('makeBadge', () => {
  it('uses the provided color', () => {
    const badge = makeBadge('Admin', 'blue');
    expect(badge.label).toBe('Admin');
    expect(badge.color).toBe('blue');
  });

  it('derives a color from the label when none provided', () => {
    const badge = makeBadge('Active');
    const validColors = ['blue', 'green', 'yellow', 'red', 'purple', 'gray', 'orange', 'teal'];
    expect(validColors).toContain(badge.color);
  });
});
