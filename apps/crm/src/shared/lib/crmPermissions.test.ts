import { describe, expect, it } from 'vitest';
import { canCrmPermission } from './crmPermissions';

describe('canCrmPermission', () => {
  it('allows owners to use every CRM action', () => {
    expect(canCrmPermission('owner', 'crm:campaigns:send')).toBe(true);
    expect(canCrmPermission('owner', 'crm:loyalty:configure')).toBe(true);
    expect(canCrmPermission('owner', 'crm:automations')).toBe(true);
  });

  it('grants admin the admin-tier CRM permissions', () => {
    expect(canCrmPermission('admin', 'crm:campaigns:send')).toBe(true);
    expect(canCrmPermission('admin', 'crm:loyalty:configure')).toBe(true);
    expect(canCrmPermission('admin', 'crm:automations')).toBe(true);
    // Admin also inherits every operator-tier permission.
    expect(canCrmPermission('admin', 'crm:customers:manage')).toBe(true);
  });

  it('grants operator the operator-tier CRM permissions directly', () => {
    expect(canCrmPermission('operator', 'crm:customers')).toBe(true);
    expect(canCrmPermission('operator', 'crm:customers:manage')).toBe(true);
    expect(canCrmPermission('operator', 'crm:campaigns')).toBe(true);
    expect(canCrmPermission('operator', 'crm:segments')).toBe(true);
    expect(canCrmPermission('operator', 'crm:loyalty')).toBe(true);
  });

  it('keeps admin-tier CRM actions away from operator', () => {
    expect(canCrmPermission('operator', 'crm:campaigns:send')).toBe(false);
    expect(canCrmPermission('operator', 'crm:loyalty:configure')).toBe(false);
    // crm:automations has no direct grant to operator in the matrix - admin-tier only.
    expect(canCrmPermission('operator', 'crm:automations')).toBe(false);
  });

  it('denies staff and viewer every CRM action since neither has app:crm', () => {
    expect(canCrmPermission('staff', 'crm:customers')).toBe(false);
    expect(canCrmPermission('staff', 'crm:campaigns')).toBe(false);
    expect(canCrmPermission('staff', 'crm:campaigns:send')).toBe(false);
    expect(canCrmPermission('viewer', 'crm:customers')).toBe(false);
    expect(canCrmPermission('viewer', 'crm:loyalty:configure')).toBe(false);
  });

  it('gates team invite/remove behind the admin-tier hub:team resources', () => {
    expect(canCrmPermission('admin', 'hub:team:invite')).toBe(true);
    expect(canCrmPermission('admin', 'hub:team:remove')).toBe(true);
    expect(canCrmPermission('operator', 'hub:team:invite')).toBe(false);
    expect(canCrmPermission('operator', 'hub:team:remove')).toBe(false);
  });

  it('denies a missing role', () => {
    expect(canCrmPermission(null, 'crm:customers')).toBe(false);
    expect(canCrmPermission(undefined, 'crm:campaigns:send')).toBe(false);
  });
});
