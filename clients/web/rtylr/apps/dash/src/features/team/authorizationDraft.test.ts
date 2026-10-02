import { describe, expect, it } from 'vitest';
import {
  buildPermissionProfileAssignments,
  getAssignedProfileIdsByApp,
} from './authorizationDraft';
import type { PermissionProfileAssignment } from '../../shared/schemas/team';

describe('buildPermissionProfileAssignments', () => {
  it.each(['explicit', 'group'] as const)(
    'preserves separate read and write profiles and their resource scopes when editing %s access',
    (assignmentSource) => {
      const existing: PermissionProfileAssignment[] = [
        {
          appId: 'flow',
          permissionProfileIds: ['read'],
          assignmentSource,
          status: 'active',
          scopes: [{ scopeType: 'flow_project', scopeId: '41' }],
        },
        {
          appId: 'flow',
          permissionProfileIds: ['write'],
          assignmentSource,
          status: 'active',
          scopes: [{ scopeType: 'flow_project', scopeId: '42' }],
        },
      ];
      const draft = getAssignedProfileIdsByApp(existing, assignmentSource);
      expect(
        buildPermissionProfileAssignments(['flow'], draft, assignmentSource, existing)
      ).toEqual(existing);

      draft.flow = ['write', 'comment'];
      expect(
        buildPermissionProfileAssignments(['flow'], draft, assignmentSource, existing)
      ).toEqual([
        existing[1],
        {
          appId: 'flow',
          permissionProfileIds: ['comment'],
          assignmentSource,
          status: 'active',
          scopes: [],
        },
      ]);
    }
  );

  it('keeps inherited group and inactive profiles out of a member direct-access draft', () => {
    const existing: PermissionProfileAssignment[] = [
      {
        appId: 'flow',
        permissionProfileIds: ['group-write'],
        assignmentSource: 'group',
        status: 'active',
        scopes: [],
      },
      {
        appId: 'flow',
        permissionProfileIds: ['old-admin'],
        assignmentSource: 'explicit',
        status: 'inactive',
        scopes: [],
      },
      {
        appId: 'flow',
        permissionProfileIds: ['read'],
        assignmentSource: 'explicit',
        status: 'active',
        scopes: [],
      },
    ];
    const draft = getAssignedProfileIdsByApp(existing, 'explicit');
    expect(buildPermissionProfileAssignments(['flow'], draft, 'explicit', existing)).toEqual([
      existing[2],
    ]);
    expect(buildPermissionProfileAssignments([], draft, 'explicit', existing)).toEqual([]);
  });

  it('creates one assignment per licensed app and removes duplicate profiles', () => {
    expect(
      buildPermissionProfileAssignments(['crm', 'flow'], {
        crm: ['profile-read', 'profile-read', 'profile-write'],
        flow: ['flow-read'],
      })
    ).toEqual([
      {
        appId: 'crm',
        permissionProfileIds: ['profile-read', 'profile-write'],
        assignmentSource: 'explicit',
        status: 'active',
        scopes: [],
      },
      {
        appId: 'flow',
        permissionProfileIds: ['flow-read'],
        assignmentSource: 'explicit',
        status: 'active',
        scopes: [],
      },
    ]);
  });

  it('omits licensed apps without a selected profile and accepts group assignments', () => {
    expect(buildPermissionProfileAssignments(['hr', 'erp'], { hr: ['hr-view'] }, 'group')).toEqual([
      {
        appId: 'hr',
        permissionProfileIds: ['hr-view'],
        assignmentSource: 'group',
        status: 'active',
        scopes: [],
      },
    ]);
  });
});
