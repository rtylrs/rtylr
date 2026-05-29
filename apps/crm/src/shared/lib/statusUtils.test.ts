import { describe, it, expect } from 'vitest';
import {
  statusLabel,
  statusColor,
  statusBadgeClass,
  isUsable,
  allowedTransitions,
  canTransition,
} from './statusUtils';

describe('statusLabel', () => {
  it('returns correct labels', () => {
    expect(statusLabel('active')).toBe('Active');
    expect(statusLabel('deleted')).toBe('Deleted');
  });
});

describe('statusColor', () => {
  it('returns color classes', () => {
    expect(statusColor('active')).toContain('green');
    expect(statusColor('deleted')).toContain('red');
  });
});

describe('statusBadgeClass', () => {
  it('returns badge classes', () => {
    expect(statusBadgeClass('active')).toContain('green-100');
    expect(statusBadgeClass('pending')).toContain('blue-100');
  });
});

describe('isUsable', () => {
  it('returns true for active and pending', () => {
    expect(isUsable('active')).toBe(true);
    expect(isUsable('pending')).toBe(true);
  });
  it('returns false for others', () => {
    expect(isUsable('inactive')).toBe(false);
    expect(isUsable('deleted')).toBe(false);
  });
});

describe('allowedTransitions', () => {
  it('returns valid next statuses', () => {
    expect(allowedTransitions('active')).toContain('inactive');
  });
  it('returns empty for deleted', () => {
    expect(allowedTransitions('deleted')).toHaveLength(0);
  });
});

describe('canTransition', () => {
  it('returns true for valid transitions', () =>
    expect(canTransition('active', 'inactive')).toBe(true));
  it('returns false for invalid', () => expect(canTransition('deleted', 'active')).toBe(false));
});
