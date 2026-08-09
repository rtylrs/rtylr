import { describe, it, expect } from 'vitest';

import {
  roleLabel,
  canManageWorkspace,
  canCreateProject,
  canEditTask,
  canDelete,
  mostRestrictiveProjectRole,
  canInviteMembers,
  rolePriority,
  sortMembersByRole,
  membersByRole,
} from './roleUtils';
import type { Member } from '../stores/flowStore';

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice',
    email: 'alice@example.com',
    initials: 'A',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

describe('roleLabel', () => {
  it('returns readable labels', () => {
    expect(roleLabel('owner')).toBe('Owner');
    expect(roleLabel('admin')).toBe('Admin');
    expect(roleLabel('member')).toBe('Member');
    expect(roleLabel('viewer')).toBe('Viewer');
  });
});

describe('canManageWorkspace', () => {
  it('returns true for owner and admin', () => {
    expect(canManageWorkspace('owner')).toBe(true);
    expect(canManageWorkspace('admin')).toBe(true);
  });

  it('returns false for member and viewer', () => {
    expect(canManageWorkspace('member')).toBe(false);
    expect(canManageWorkspace('viewer')).toBe(false);
  });
});

describe('canCreateProject', () => {
  it('returns true for owner, admin, member', () => {
    expect(canCreateProject('owner')).toBe(true);
    expect(canCreateProject('admin')).toBe(true);
    expect(canCreateProject('member')).toBe(true);
  });

  it('returns false for viewer', () => {
    expect(canCreateProject('viewer')).toBe(false);
  });
});

describe('canEditTask', () => {
  it('returns true for non-viewer project roles', () => {
    expect(canEditTask('member')).toBe(true);
    expect(canEditTask('admin')).toBe(true);
  });

  it('returns false for a project viewer', () => {
    expect(canEditTask('viewer')).toBe(false);
  });
});

describe('canDelete', () => {
  it('returns true for a project admin', () => {
    expect(canDelete('admin')).toBe(true);
  });

  it('returns false for project members and viewers', () => {
    expect(canDelete('member')).toBe(false);
    expect(canDelete('viewer')).toBe(false);
  });
});

describe('mostRestrictiveProjectRole', () => {
  it('returns undefined for an empty list', () => {
    expect(mostRestrictiveProjectRole([])).toBeUndefined();
  });

  it('returns the only role for a single project', () => {
    expect(mostRestrictiveProjectRole(['admin'])).toBe('admin');
  });

  it('picks the least privileged role across projects', () => {
    expect(mostRestrictiveProjectRole(['admin', 'viewer'])).toBe('viewer');
    expect(mostRestrictiveProjectRole(['admin', 'member'])).toBe('member');
    expect(mostRestrictiveProjectRole(['member', 'viewer', 'admin'])).toBe('viewer');
  });

  it('is order independent', () => {
    expect(mostRestrictiveProjectRole(['viewer', 'admin'])).toBe(
      mostRestrictiveProjectRole(['admin', 'viewer'])
    );
  });

  it('denies editing when one project in the selection is view-only', () => {
    const role = mostRestrictiveProjectRole(['admin', 'admin', 'viewer']);
    expect(role !== undefined && canEditTask(role)).toBe(false);
  });
});

describe('canInviteMembers', () => {
  it('returns true for owner and admin', () => {
    expect(canInviteMembers('owner')).toBe(true);
  });

  it('returns false for member', () => {
    expect(canInviteMembers('member')).toBe(false);
  });
});

describe('rolePriority', () => {
  it('owner has lowest priority number', () => {
    expect(rolePriority('owner')).toBeLessThan(rolePriority('viewer'));
  });
});

describe('sortMembersByRole', () => {
  it('sorts owners first', () => {
    const members = [makeMember({ role: 'viewer' }), makeMember({ id: 'm2', role: 'owner' })];
    expect(sortMembersByRole(members)[0].role).toBe('owner');
  });
});

describe('membersByRole', () => {
  it('filters by role', () => {
    const members = [makeMember({ role: 'owner' }), makeMember({ id: 'm2', role: 'member' })];
    expect(membersByRole(members, 'owner').length).toBe(1);
  });
});
