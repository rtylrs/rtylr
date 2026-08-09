import { describe, it, expect } from 'vitest';
import {
  allStatuses,
  statusColor,
  statusLabel,
  statusIndex,
  isTerminalStatus,
  isActiveStatus,
} from './status';

describe('allStatuses', () => {
  it('returns all 7 statuses', () => {
    expect(allStatuses()).toHaveLength(7);
  });

  it('starts with new', () => {
    expect(allStatuses()[0]).toBe('new');
  });
});

describe('statusColor', () => {
  it('returns correct colors', () => {
    expect(statusColor('new')).toBe('blue');
    expect(statusColor('hired')).toBe('green');
    expect(statusColor('rejected')).toBe('red');
  });

  it('returns gray for unknown status', () => {
    expect(statusColor('unknown')).toBe('gray');
  });
});

describe('statusLabel', () => {
  it('returns human-readable labels', () => {
    expect(statusLabel('new')).toBe('New');
    expect(statusLabel('offer')).toBe('Offer Made');
  });

  it('returns the status for unknown values', () => {
    expect(statusLabel('custom')).toBe('custom');
  });
});

describe('statusIndex', () => {
  it('returns 0 for new', () => {
    expect(statusIndex('new')).toBe(0);
  });

  it('returns -1 for unknown', () => {
    expect(statusIndex('nonexistent')).toBe(-1);
  });
});

describe('isTerminalStatus', () => {
  it('returns true for terminal statuses', () => {
    expect(isTerminalStatus('hired')).toBe(true);
    expect(isTerminalStatus('rejected')).toBe(true);
    expect(isTerminalStatus('withdrawn')).toBe(true);
  });

  it('returns false for active statuses', () => {
    expect(isTerminalStatus('new')).toBe(false);
  });
});

describe('isActiveStatus', () => {
  it('returns true for non-terminal statuses', () => {
    expect(isActiveStatus('new')).toBe(true);
    expect(isActiveStatus('interview')).toBe(true);
  });

  it('returns false for terminal statuses', () => {
    expect(isActiveStatus('hired')).toBe(false);
  });
});
