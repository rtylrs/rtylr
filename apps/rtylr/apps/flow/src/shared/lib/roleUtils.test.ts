import { describe, it, expect } from 'vitest';

import {
  roleLabel,
  canManageWorkspace,
  canCreateProject,
  canEditTask,
  canDelete,
  hasFlowPermission,
  canFlowPermission,
  mostRestrictiveProjectRole,
  canInviteMembers,
  toFlowMemberRole,
  rolePriority,
  sortMembersByRole,
  membersByRole,
} from './roleUtils';
import type { Member } from '../stores/flowStore';
import type { EffectiveAccess } from '@voxire/auth';

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
  it('returns a translation key for each role, not a hardcoded string', () => {
    expect(roleLabel('owner')).toBe('common.roles.owner');
    expect(roleLabel('admin')).toBe('common.roles.admin');
    expect(roleLabel('member')).toBe('common.roles.member');
    expect(roleLabel('viewer')).toBe('common.roles.viewer');
  });

  it('returns a distinct key per role', () => {
    const roles: Array<Member['role']> = ['owner', 'admin', 'member', 'viewer'];
    const keys = new Set(roles.map(roleLabel));
    expect(keys.size).toBe(roles.length);
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
    expect(canInviteMembers('admin')).toBe(true);
  });

  it('returns false for member and viewer', () => {
    expect(canInviteMembers('member')).toBe(false);
    expect(canInviteMembers('viewer')).toBe(false);
  });
});

describe('toFlowMemberRole', () => {
  it('passes through the four roles Flow knows', () => {
    expect(toFlowMemberRole('owner')).toBe('owner');
    expect(toFlowMemberRole('admin')).toBe('admin');
    expect(toFlowMemberRole('member')).toBe('member');
    expect(toFlowMemberRole('viewer')).toBe('viewer');
  });

  it('defaults an unrecognized business-wide role to viewer, not member', () => {
    // operator/staff/billing/menu_manager/store_manager are real business-wide
    // roles for non-Flow business types - none of them imply Flow permissions.
    expect(toFlowMemberRole('operator')).toBe('viewer');
    expect(toFlowMemberRole('staff')).toBe('viewer');
    expect(toFlowMemberRole('billing')).toBe('viewer');
    expect(toFlowMemberRole('menu_manager')).toBe('viewer');
    expect(toFlowMemberRole('store_manager')).toBe('viewer');
    expect(toFlowMemberRole('')).toBe('viewer');
  });

  it('an unrecognized role therefore cannot create projects or invite members', () => {
    const role = toFlowMemberRole('operator');
    expect(canCreateProject(role)).toBe(false);
    expect(canInviteMembers(role)).toBe(false);
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

describe('hasFlowPermission', () => {
  const access: EffectiveAccess = {
    businessId: 'b1',
    dashRole: 'member',
    dashRoleId: null,
    dashPermissions: [],
    subscribedApps: ['flow'],
    licenses: [],
    isOwner: false,
    isAdmin: false,
    authorizationVersion: 1,
    apps: [
      {
        appId: 'flow',
        licensed: true,
        hasAccess: true,
        roleIds: [],
        primaryRoleId: null,
        displayRole: null,
        permissions: ['flow:projects:view'],
        scopes: [],
        sources: [],
        permissionSources: {},
      },
    ],
  };

  it('requires a licensed Flow app with the requested permission', () => {
    expect(hasFlowPermission(access, 'flow:projects:view')).toBe(true);
    expect(hasFlowPermission(access, 'flow:settings:manage')).toBe(false);
  });

  it('fails closed when effective access is absent or the app is unlicensed', () => {
    expect(hasFlowPermission(null, 'flow:projects:view')).toBe(false);
    expect(
      hasFlowPermission(
        {
          ...access,
          apps: [{ ...access.apps[0], licensed: false }],
        },
        'flow:projects:view'
      )
    ).toBe(false);
  });
});

describe('canFlowPermission legacy role fallback', () => {
  const flowApp = {
    appId: 'flow',
    licensed: true,
    hasAccess: true,
    roleIds: [],
    primaryRoleId: null,
    displayRole: null,
    permissions: ['flow:projects:view'],
    scopes: [],
    sources: [],
    permissionSources: {},
  };
  const access: EffectiveAccess = {
    businessId: 'b1',
    dashRole: 'member',
    dashRoleId: null,
    dashPermissions: [],
    subscribedApps: ['flow'],
    licenses: [],
    isOwner: false,
    isAdmin: false,
    authorizationVersion: 1,
    apps: [flowApp],
  };

  // The regression: a getMe payload that carries a role and no effectiveAccess.
  describe('with no effective access in the payload', () => {
    it.each(['owner', 'admin'])('lets %s manage settings and the team', (role) => {
      expect(canFlowPermission(role, 'flow:settings:manage', undefined)).toBe(true);
      expect(canFlowPermission(role, 'flow:team:manage', undefined)).toBe(true);
      expect(canFlowPermission(role, 'flow:projects:create', undefined)).toBe(true);
    });

    it('lets a member create a project but not manage settings or the team', () => {
      expect(canFlowPermission('member', 'flow:projects:create', undefined)).toBe(true);
      expect(canFlowPermission('member', 'flow:settings:manage', undefined)).toBe(false);
      expect(canFlowPermission('member', 'flow:team:manage', undefined)).toBe(false);
    });

    it('grants a viewer nothing', () => {
      expect(canFlowPermission('viewer', 'flow:projects:create', undefined)).toBe(false);
      expect(canFlowPermission('viewer', 'flow:settings:manage', undefined)).toBe(false);
    });

    it('treats a role Flow does not recognise as a viewer', () => {
      expect(canFlowPermission('store_manager', 'flow:projects:create', undefined)).toBe(false);
      expect(canFlowPermission('billing', 'flow:settings:manage', undefined)).toBe(false);
    });

    it.each([null, undefined, ''])('denies when the role is %s too', (role) => {
      expect(canFlowPermission(role, 'flow:projects:create', undefined)).toBe(false);
    });
  });

  // When the payload is present it is authoritative, and the legacy role must
  // not widen it: that would make the fallback a permanent privilege escalation
  // rather than a migration bridge.
  describe('with effective access in the payload', () => {
    it('answers from the payload and ignores the legacy role', () => {
      expect(canFlowPermission('owner', 'flow:projects:create', access)).toBe(false);
      expect(canFlowPermission('viewer', 'flow:projects:create', access)).toBe(false);
    });

    it('grants what the payload grants', () => {
      const granted: EffectiveAccess = {
        ...access,
        apps: [{ ...flowApp, permissions: ['flow:projects:create'] }],
      };
      expect(canFlowPermission('viewer', 'flow:projects:create', granted)).toBe(true);
    });
  });
});
