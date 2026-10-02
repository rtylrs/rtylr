import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canAccessErpPath, getErpHomePath } from './navigation';

const member = (permissions: string[]) => ({
  effectiveAccess: createEffectiveAccess('erp', { permissions }),
});

describe('ERP navigation inside Portal', () => {
  it('opens the catalog for a products-only member without exposing inventory or purchasing', () => {
    const user = member(['erp:products:view']);
    expect(getErpHomePath(user)).toBe('/erp/products');
    expect(canAccessErpPath(user, '/erp/products/42')).toBe(true);
    expect(canAccessErpPath(user, '/erp/categories')).toBe(true);
    expect(canAccessErpPath(user, '/erp/inventory')).toBe(false);
    expect(canAccessErpPath(user, '/erp/purchasing')).toBe(false);
    expect(canAccessErpPath(user, '/settings/erp')).toBe(false);
  });
  it('opens purchasing for a purchasing-only member and keeps warehouse management separate', () => {
    const user = member(['erp:purchasing:view']);
    expect(getErpHomePath(user)).toBe('/erp/receiving');
    expect(canAccessErpPath(user, '/erp/purchasing/42')).toBe(true);
    expect(canAccessErpPath(user, '/erp/warehouses')).toBe(false);
  });
  it('requires an active app grant for ERP deep links and settings', () => {
    expect(
      canAccessErpPath(
        { effectiveAccess: createEffectiveAccess('erp', { licensed: false }) },
        '/erp/products'
      )
    ).toBe(false);
    expect(
      canAccessErpPath(
        { effectiveAccess: createEffectiveAccess('crm', { isOwner: true }) },
        '/settings/erp'
      )
    ).toBe(false);
  });
});
