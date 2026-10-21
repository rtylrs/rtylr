import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canFinancePermission } from './financePermissions';

describe('canFinancePermission', () => {
  it('allows the owner every action in a subscribed Finance app', () => {
    const access = createEffectiveAccess('finance', { isOwner: true });
    expect(canFinancePermission('finance:expenses:approve', access)).toBe(true);
    expect(canFinancePermission('finance:invoices:send', access)).toBe(true);
    expect(canFinancePermission('finance:quotes:manage', access)).toBe(true);
    expect(canFinancePermission('finance:recurring_invoices:manage', access)).toBe(true);
    expect(canFinancePermission('hub:team:invite', access)).toBe(true);
  });

  it('allows an admin every action in a subscribed Finance app', () => {
    const access = createEffectiveAccess('finance', { isAdmin: true });
    expect(canFinancePermission('finance:expenses:approve', access)).toBe(true);
    expect(canFinancePermission('finance:invoices:send', access)).toBe(true);
    expect(canFinancePermission('finance:quotes:manage', access)).toBe(true);
    expect(canFinancePermission('finance:recurring_invoices:manage', access)).toBe(true);
    expect(canFinancePermission('hub:team:invite', access)).toBe(true);
  });

  it('allows only assigned permissions for a regular member', () => {
    const access = createEffectiveAccess('finance', {
      permissions: ['finance:quotes:manage', 'finance:recurring_invoices:manage'],
    });
    expect(canFinancePermission('finance:quotes:manage', access)).toBe(true);
    expect(canFinancePermission('finance:recurring_invoices:manage', access)).toBe(true);
    expect(canFinancePermission('finance:expenses:approve', access)).toBe(false);
    expect(canFinancePermission('hub:team:invite', access)).toBe(false);
  });

  it('denies permissions without effective access data', () => {
    expect(canFinancePermission('finance:tax:configure', null)).toBe(false);
    expect(canFinancePermission('finance:accounts:view', undefined)).toBe(false);
  });
});
