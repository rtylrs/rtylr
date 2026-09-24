import { describe, expect, it } from 'vitest';
import { canFinancePermission } from './financePermissions';

describe('canFinancePermission', () => {
  it('allows owners to use every Finance action', () => {
    expect(canFinancePermission('owner', 'finance:expenses:approve')).toBe(true);
    expect(canFinancePermission('owner', 'finance:invoices:send')).toBe(true);
    expect(canFinancePermission('owner', 'finance:quotes:manage')).toBe(true);
    expect(canFinancePermission('owner', 'finance:recurring_invoices:manage')).toBe(true);
    expect(canFinancePermission('owner', 'finance:tax')).toBe(true);
    expect(canFinancePermission('owner', 'finance:accounts')).toBe(true);
    expect(canFinancePermission('owner', 'finance:reconcile')).toBe(true);
    expect(canFinancePermission('owner', 'hub:team:invite')).toBe(true);
    expect(canFinancePermission('owner', 'hub:team:remove')).toBe(true);
  });

  it('allows admin to use every admin-tier Finance action', () => {
    expect(canFinancePermission('admin', 'finance:expenses:approve')).toBe(true);
    expect(canFinancePermission('admin', 'finance:invoices:send')).toBe(true);
    expect(canFinancePermission('admin', 'finance:quotes:manage')).toBe(true);
    expect(canFinancePermission('admin', 'finance:recurring_invoices:manage')).toBe(true);
    expect(canFinancePermission('admin', 'finance:tax')).toBe(true);
    expect(canFinancePermission('admin', 'finance:accounts')).toBe(true);
    expect(canFinancePermission('admin', 'finance:reconcile')).toBe(true);
    expect(canFinancePermission('admin', 'hub:team:invite')).toBe(true);
    expect(canFinancePermission('admin', 'hub:team:remove')).toBe(true);
  });

  it('allows the billing role to manage quotes and recurring invoices', () => {
    expect(canFinancePermission('billing', 'finance:quotes:manage')).toBe(true);
    expect(canFinancePermission('billing', 'finance:recurring_invoices:manage')).toBe(true);
  });

  it('keeps admin-tier Finance actions away from operator', () => {
    expect(canFinancePermission('operator', 'finance:expenses:approve')).toBe(false);
    expect(canFinancePermission('operator', 'finance:invoices:send')).toBe(false);
    expect(canFinancePermission('operator', 'finance:quotes:manage')).toBe(false);
    expect(canFinancePermission('operator', 'finance:recurring_invoices:manage')).toBe(false);
    expect(canFinancePermission('operator', 'finance:tax')).toBe(false);
    expect(canFinancePermission('operator', 'finance:accounts')).toBe(false);
    expect(canFinancePermission('operator', 'finance:reconcile')).toBe(false);
    expect(canFinancePermission('operator', 'hub:team:invite')).toBe(false);
    expect(canFinancePermission('operator', 'hub:team:remove')).toBe(false);
  });

  it('keeps admin-tier Finance actions away from staff', () => {
    expect(canFinancePermission('staff', 'finance:expenses:approve')).toBe(false);
    expect(canFinancePermission('staff', 'finance:invoices:send')).toBe(false);
    expect(canFinancePermission('staff', 'finance:quotes:manage')).toBe(false);
    expect(canFinancePermission('staff', 'finance:recurring_invoices:manage')).toBe(false);
    expect(canFinancePermission('staff', 'finance:tax')).toBe(false);
    expect(canFinancePermission('staff', 'finance:accounts')).toBe(false);
    expect(canFinancePermission('staff', 'finance:reconcile')).toBe(false);
  });

  it('keeps admin-tier Finance actions away from viewer', () => {
    expect(canFinancePermission('viewer', 'finance:expenses:approve')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:invoices:send')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:quotes:manage')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:recurring_invoices:manage')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:tax')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:accounts')).toBe(false);
    expect(canFinancePermission('viewer', 'finance:reconcile')).toBe(false);
  });

  it('denies a missing role', () => {
    expect(canFinancePermission(null, 'finance:tax')).toBe(false);
    expect(canFinancePermission(undefined, 'finance:accounts')).toBe(false);
  });
});
