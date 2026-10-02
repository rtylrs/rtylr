import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canAccessPortalPath } from './canAccessPortalPath';

describe('Portal migrated app deep links', () => {
  it('allows legacy CRM home redirects for a campaigns-only member without widening customer access', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('crm', { permissions: ['crm:campaigns:view'] }),
    };
    expect(canAccessPortalPath(user, '/crm/')).toBe(true);
    expect(canAccessPortalPath(user, '/crm/campaigns/')).toBe(true);
    expect(canAccessPortalPath(user, '/crm/customers/')).toBe(false);
  });
  it('allows legacy ERP home and settings redirects while preserving the member permission boundary', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('erp', {
        permissions: ['erp:products:view', 'erp:settings:view'],
      }),
    };
    expect(canAccessPortalPath(user, '/erp/')).toBe(true);
    expect(canAccessPortalPath(user, '/settings/erp/')).toBe(true);
    expect(canAccessPortalPath(user, '/erp/inventory/')).toBe(false);
    expect(canAccessPortalPath(user, '/settings/crm/')).toBe(false);
  });
});
