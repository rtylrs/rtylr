import { describe, expect, it } from 'vitest';
import { canErpPermission } from './erpPermissions';

describe('canErpPermission', () => {
  it('allows owners to use every ERP action', () => {
    expect(canErpPermission('owner', 'erp:products:delete')).toBe(true);
    expect(canErpPermission('owner', 'erp:inventory:adjust')).toBe(true);
    expect(canErpPermission('owner', 'erp:inventory:transfer')).toBe(true);
    expect(canErpPermission('owner', 'erp:purchase_orders:approve')).toBe(true);
    expect(canErpPermission('owner', 'erp:stock_takes')).toBe(true);
    expect(canErpPermission('owner', 'erp:settings:manage')).toBe(true);
    expect(canErpPermission('owner', 'erp:suppliers:create')).toBe(true);
    expect(canErpPermission('owner', 'hub:team:invite')).toBe(true);
    expect(canErpPermission('owner', 'hub:team:remove')).toBe(true);
  });

  it('allows admin to use every admin-tier ERP action', () => {
    expect(canErpPermission('admin', 'erp:products:delete')).toBe(true);
    expect(canErpPermission('admin', 'erp:inventory:adjust')).toBe(true);
    expect(canErpPermission('admin', 'erp:inventory:transfer')).toBe(true);
    expect(canErpPermission('admin', 'erp:purchase_orders:approve')).toBe(true);
    expect(canErpPermission('admin', 'erp:stock_takes')).toBe(true);
    expect(canErpPermission('admin', 'erp:settings:manage')).toBe(true);
    expect(canErpPermission('admin', 'erp:suppliers:create')).toBe(true);
    expect(canErpPermission('admin', 'hub:team:invite')).toBe(true);
    expect(canErpPermission('admin', 'hub:team:remove')).toBe(true);
  });

  it('grants an operator the view/create/edit tier but keeps admin-tier ERP actions away', () => {
    expect(canErpPermission('operator', 'erp:products')).toBe(true);
    expect(canErpPermission('operator', 'erp:products:create')).toBe(true);
    expect(canErpPermission('operator', 'erp:products:edit')).toBe(true);
    expect(canErpPermission('operator', 'erp:inventory')).toBe(true);
    expect(canErpPermission('operator', 'erp:suppliers')).toBe(true);
    expect(canErpPermission('operator', 'erp:purchase_orders')).toBe(true);

    expect(canErpPermission('operator', 'erp:suppliers:create')).toBe(false);
    expect(canErpPermission('operator', 'erp:settings:manage')).toBe(false);
    expect(canErpPermission('operator', 'erp:products:delete')).toBe(false);
    expect(canErpPermission('operator', 'erp:inventory:adjust')).toBe(false);
    expect(canErpPermission('operator', 'erp:inventory:transfer')).toBe(false);
    expect(canErpPermission('operator', 'erp:purchase_orders:approve')).toBe(false);
    expect(canErpPermission('operator', 'erp:stock_takes')).toBe(false);
    expect(canErpPermission('operator', 'hub:team:invite')).toBe(false);
    expect(canErpPermission('operator', 'hub:team:remove')).toBe(false);
  });

  it('keeps every ERP action away from staff and viewer, who have no ERP grant', () => {
    expect(canErpPermission('staff', 'erp:products')).toBe(false);
    expect(canErpPermission('staff', 'erp:products:delete')).toBe(false);
    expect(canErpPermission('staff', 'erp:inventory:adjust')).toBe(false);
    expect(canErpPermission('staff', 'erp:purchase_orders:approve')).toBe(false);
    expect(canErpPermission('staff', 'hub:team:invite')).toBe(false);

    expect(canErpPermission('viewer', 'erp:products')).toBe(false);
    expect(canErpPermission('viewer', 'erp:products:delete')).toBe(false);
    expect(canErpPermission('viewer', 'erp:inventory:adjust')).toBe(false);
    expect(canErpPermission('viewer', 'erp:purchase_orders:approve')).toBe(false);
    expect(canErpPermission('viewer', 'hub:team:remove')).toBe(false);
  });

  it('denies a missing role', () => {
    expect(canErpPermission(null, 'erp:products:delete')).toBe(false);
    expect(canErpPermission(undefined, 'erp:inventory:adjust')).toBe(false);
  });
});
