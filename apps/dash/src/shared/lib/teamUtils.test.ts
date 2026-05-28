import { describe, it, expect } from 'vitest';
import {
  roleLabel,
  isAdminRole,
  canManageUsers,
  sortTeamMembers,
  activeMembers,
  findByEmail,
  countByRole,
} from './teamUtils';
import type { TeamMember } from './teamUtils';

const makeMembers = (): TeamMember[] => [
  { id: '1', name: 'Alice', email: 'alice@example.com', role: 'owner', active: true },
  { id: '2', name: 'Bob', email: 'bob@example.com', role: 'staff', active: true },
  { id: '3', name: 'Carol', email: 'carol@example.com', role: 'admin', active: false },
  { id: '4', name: 'Dave', email: 'dave@example.com', role: 'manager', active: true },
];

describe('roleLabel', () => {
  it('returns Owner for owner role', () => {
    expect(roleLabel('owner')).toBe('Owner');
  });

  it('returns Admin for admin role', () => {
    expect(roleLabel('admin')).toBe('Admin');
  });

  it('returns Manager for manager role', () => {
    expect(roleLabel('manager')).toBe('Manager');
  });

  it('returns Staff for staff role', () => {
    expect(roleLabel('staff')).toBe('Staff');
  });

  it('returns Viewer for viewer role', () => {
    expect(roleLabel('viewer')).toBe('Viewer');
  });
});

describe('isAdminRole', () => {
  it('returns true for owner', () => {
    expect(isAdminRole('owner')).toBe(true);
  });

  it('returns true for admin', () => {
    expect(isAdminRole('admin')).toBe(true);
  });

  it('returns false for manager', () => {
    expect(isAdminRole('manager')).toBe(false);
  });

  it('returns false for staff', () => {
    expect(isAdminRole('staff')).toBe(false);
  });

  it('returns false for viewer', () => {
    expect(isAdminRole('viewer')).toBe(false);
  });
});

describe('canManageUsers', () => {
  it('returns true for owner', () => {
    expect(canManageUsers('owner')).toBe(true);
  });

  it('returns true for admin', () => {
    expect(canManageUsers('admin')).toBe(true);
  });

  it('returns true for manager', () => {
    expect(canManageUsers('manager')).toBe(true);
  });

  it('returns false for staff', () => {
    expect(canManageUsers('staff')).toBe(false);
  });

  it('returns false for viewer', () => {
    expect(canManageUsers('viewer')).toBe(false);
  });
});

describe('sortTeamMembers', () => {
  it('sorts owners before admins before managers', () => {
    const members = makeMembers();
    const sorted = sortTeamMembers(members);
    expect(sorted[0].role).toBe('owner');
    expect(sorted[1].role).toBe('admin');
    expect(sorted[2].role).toBe('manager');
  });

  it('does not mutate the original array', () => {
    const members = makeMembers();
    sortTeamMembers(members);
    expect(members[0].id).toBe('1');
  });

  it('sorts alphabetically within the same role', () => {
    const members: TeamMember[] = [
      { id: '2', name: 'Zara', email: 'z@x.com', role: 'staff' },
      { id: '1', name: 'Aaron', email: 'a@x.com', role: 'staff' },
    ];
    const sorted = sortTeamMembers(members);
    expect(sorted[0].name).toBe('Aaron');
  });
});

describe('activeMembers', () => {
  it('returns only active members', () => {
    const result = activeMembers(makeMembers());
    expect(result.every((m) => m.active !== false)).toBe(true);
  });

  it('excludes members with active=false', () => {
    const members = makeMembers();
    const result = activeMembers(members);
    expect(result.find((m) => m.id === '3')).toBeUndefined();
  });

  it('includes members without explicit active field', () => {
    const members: TeamMember[] = [{ id: '1', name: 'X', email: 'x@x.com', role: 'staff' }];
    expect(activeMembers(members)).toHaveLength(1);
  });
});

describe('findByEmail', () => {
  it('finds a member by exact email', () => {
    const result = findByEmail(makeMembers(), 'alice@example.com');
    expect(result?.id).toBe('1');
  });

  it('is case-insensitive', () => {
    const result = findByEmail(makeMembers(), 'ALICE@EXAMPLE.COM');
    expect(result?.id).toBe('1');
  });

  it('returns undefined when not found', () => {
    expect(findByEmail(makeMembers(), 'nobody@example.com')).toBeUndefined();
  });
});

describe('countByRole', () => {
  it('counts members by role', () => {
    const counts = countByRole(makeMembers());
    expect(counts.owner).toBe(1);
    expect(counts.admin).toBe(1);
    expect(counts.manager).toBe(1);
    expect(counts.staff).toBe(1);
    expect(counts.viewer).toBe(0);
  });

  it('returns zeros for all roles with empty input', () => {
    const counts = countByRole([]);
    expect(Object.values(counts).every((v) => v === 0)).toBe(true);
  });
});
