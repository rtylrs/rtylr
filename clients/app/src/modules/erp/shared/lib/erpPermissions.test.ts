import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canErpPermission } from './erpPermissions';

describe('canErpPermission', () => {
  it('allows owners and admins every action in a subscribed ERP app', () => {
    for (const access of [
      createEffectiveAccess('erp', { isOwner: true }),
      createEffectiveAccess('erp', { isAdmin: true }),
    ]) {
      expect(canErpPermission('erp:products:delete', access)).toBe(true);
      expect(canErpPermission('erp:inventory:adjust', access)).toBe(true);
      expect(canErpPermission('erp:purchasing:approve', access)).toBe(true);
      expect(canErpPermission('hub:team:invite', access)).toBe(true);
    }
  });

  it('allows a member only the ERP permissions assigned to their profiles', () => {
    const access = createEffectiveAccess('erp', {
      permissions: ['erp:products:view', 'erp:products:create', 'erp:products:edit'],
    });
    expect(canErpPermission('erp:products:view', access)).toBe(true);
    expect(canErpPermission('erp:products:create', access)).toBe(true);
    expect(canErpPermission('erp:products:edit', access)).toBe(true);
    expect(canErpPermission('erp:products:delete', access)).toBe(false);
    expect(canErpPermission('erp:purchasing:approve', access)).toBe(false);
  });

  it('denies permissions when effective access is missing', () => {
    expect(canErpPermission('erp:products:delete', null)).toBe(false);
    expect(canErpPermission('erp:inventory:adjust', undefined)).toBe(false);
  });
});
