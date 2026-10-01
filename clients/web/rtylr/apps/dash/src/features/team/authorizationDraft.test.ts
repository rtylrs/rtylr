import { describe, expect, it } from 'vitest';
import { buildPermissionProfileAssignments } from './authorizationDraft';

describe('buildPermissionProfileAssignments', () => {
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
