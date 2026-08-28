import { describe, it, expect } from 'vitest';
import type { EffectiveAccess } from '@voxire/auth';
import { canPurgeProject } from './canPurgeProject';

const project = { createdBy: '2001', ownerId: '3003' };
const flowManagerAccess: EffectiveAccess = {
  businessId: 'b1',
  dashRole: 'member',
  dashRoleId: null,
  dashPermissions: [],
  subscribedApps: ['flow'],
  licenses: [],
  apps: [
    {
      appId: 'flow',
      licensed: true,
      hasAccess: true,
      roleIds: [],
      primaryRoleId: null,
      displayRole: null,
      permissions: ['flow:settings:manage'],
      scopes: [],
      sources: [],
      permissionSources: {},
    },
  ],
  isOwner: false,
  isAdmin: false,
  authorizationVersion: 1,
};

describe('canPurgeProject', () => {
  it('admits the creator', () => {
    expect(canPurgeProject(project, '2001', null)).toBe(true);
  });

  it('admits the project owner', () => {
    expect(canPurgeProject(project, '3003', null)).toBe(true);
  });

  it('admits a central Flow settings manager', () => {
    expect(canPurgeProject(project, '9999', flowManagerAccess)).toBe(true);
  });

  it('denies a user without central permission when they neither created nor own the project', () => {
    expect(canPurgeProject(project, '9999', null)).toBe(false);
  });

  it('denies an unknown user id', () => {
    expect(canPurgeProject(project, null, null)).toBe(false);
    expect(canPurgeProject(project, '', null)).toBe(false);
  });

  it('denies a null role that is not the creator or owner', () => {
    expect(canPurgeProject(project, '9999', null)).toBe(false);
  });

  // An absent id must not match an absent user id into a false positive.
  it('denies a project whose creator and owner are unset', () => {
    expect(canPurgeProject({}, '2001', null)).toBe(false);
  });
});
