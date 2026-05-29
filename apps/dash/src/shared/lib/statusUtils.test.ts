import { describe, it, expect } from 'vitest';
import {
  statusLabel,
  statusVariant,
  isActiveStatus,
  isTerminalStatus,
  allowedTransitions,
  canTransition,
  activeStatuses,
} from './statusUtils';

describe('statusLabel', () => {
  it('returns "Active" for active', () => {
    expect(statusLabel('active')).toBe('Active');
  });

  it('returns "Pending" for pending', () => {
    expect(statusLabel('pending')).toBe('Pending');
  });

  it('returns "Archived" for archived', () => {
    expect(statusLabel('archived')).toBe('Archived');
  });
});

describe('statusVariant', () => {
  it('returns success for active', () => {
    expect(statusVariant('active')).toBe('success');
  });

  it('returns warning for pending', () => {
    expect(statusVariant('pending')).toBe('warning');
  });

  it('returns error for suspended', () => {
    expect(statusVariant('suspended')).toBe('error');
  });

  it('returns neutral for inactive', () => {
    expect(statusVariant('inactive')).toBe('neutral');
  });
});

describe('isActiveStatus', () => {
  it('returns true for active', () => {
    expect(isActiveStatus('active')).toBe(true);
  });

  it('returns false for inactive', () => {
    expect(isActiveStatus('inactive')).toBe(false);
  });

  it('returns false for pending', () => {
    expect(isActiveStatus('pending')).toBe(false);
  });
});

describe('isTerminalStatus', () => {
  it('returns true for archived', () => {
    expect(isTerminalStatus('archived')).toBe(true);
  });

  it('returns true for suspended', () => {
    expect(isTerminalStatus('suspended')).toBe(true);
  });

  it('returns false for active', () => {
    expect(isTerminalStatus('active')).toBe(false);
  });

  it('returns false for pending', () => {
    expect(isTerminalStatus('pending')).toBe(false);
  });
});

describe('allowedTransitions', () => {
  it('returns transitions from active', () => {
    expect(allowedTransitions('active')).toContain('inactive');
    expect(allowedTransitions('active')).toContain('suspended');
  });

  it('returns empty array for archived (terminal)', () => {
    expect(allowedTransitions('archived')).toEqual([]);
  });

  it('returns active from inactive', () => {
    expect(allowedTransitions('inactive')).toContain('active');
  });
});

describe('canTransition', () => {
  it('returns true for allowed transitions', () => {
    expect(canTransition('active', 'inactive')).toBe(true);
    expect(canTransition('pending', 'active')).toBe(true);
  });

  it('returns false for disallowed transitions', () => {
    expect(canTransition('archived', 'active')).toBe(false);
    expect(canTransition('active', 'pending')).toBe(false);
  });
});

describe('activeStatuses', () => {
  it('includes active, inactive, and pending', () => {
    const statuses = activeStatuses();
    expect(statuses).toContain('active');
    expect(statuses).toContain('inactive');
    expect(statuses).toContain('pending');
  });

  it('does not include terminal statuses', () => {
    const statuses = activeStatuses();
    expect(statuses).not.toContain('archived');
    expect(statuses).not.toContain('suspended');
  });
});
