import { describe, it, expect } from 'vitest';
import {
  roleLabel,
  roleDescription,
  outranks,
  atLeastRole,
  allRoles,
  managedBy,
  roleBadgeColor,
} from './roleUtils';

describe('roleLabel', () => {
  it('returns Owner for owner', () => expect(roleLabel('owner')).toBe('Owner'));
  it('returns Kitchen Staff for kitchen', () => expect(roleLabel('kitchen')).toBe('Kitchen Staff'));
  it('returns Runner for runner', () => expect(roleLabel('runner')).toBe('Runner'));
});

describe('roleDescription', () => {
  it('returns a non-empty description for each role', () => {
    const roles = ['owner', 'manager', 'cashier', 'server', 'kitchen', 'host', 'runner'] as const;
    for (const role of roles) {
      expect(roleDescription(role).length).toBeGreaterThan(0);
    }
  });
});

describe('outranks', () => {
  it('returns true when owner outranks manager', () => {
    expect(outranks('owner', 'manager')).toBe(true);
  });

  it('returns false when runner outranks owner', () => {
    expect(outranks('runner', 'owner')).toBe(false);
  });

  it('returns false for same role', () => {
    expect(outranks('cashier', 'cashier')).toBe(false);
  });
});

describe('atLeastRole', () => {
  it('returns true when role equals threshold', () => {
    expect(atLeastRole('manager', 'manager')).toBe(true);
  });

  it('returns true when role outranks threshold', () => {
    expect(atLeastRole('owner', 'cashier')).toBe(true);
  });

  it('returns false when role is below threshold', () => {
    expect(atLeastRole('runner', 'manager')).toBe(false);
  });
});

describe('allRoles', () => {
  it('returns all 7 roles', () => {
    expect(allRoles()).toHaveLength(7);
  });

  it('starts with owner', () => {
    expect(allRoles()[0]).toBe('owner');
  });
});

describe('managedBy', () => {
  it('returns no roles for owner', () => {
    expect(managedBy('owner')).toHaveLength(0);
  });

  it('returns owner as manager of manager', () => {
    expect(managedBy('manager')).toEqual(['owner']);
  });

  it('returns owner and manager for cashier', () => {
    expect(managedBy('cashier')).toEqual(['owner', 'manager']);
  });
});

describe('roleBadgeColor', () => {
  it('returns purple for owner', () => expect(roleBadgeColor('owner')).toBe('purple'));
  it('returns gray for runner', () => expect(roleBadgeColor('runner')).toBe('gray'));
});
