import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../test/effectiveAccess';
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

describe('remaining app member routes', () => {
  it('keeps Finance invoices readable without granting creation, expenses or settings', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('finance', { permissions: ['finance:invoices:view'] }),
    };
    expect(canAccessPortalPath(user, '/finance/invoices')).toBe(true);
    expect(canAccessPortalPath(user, '/finance/invoices/123')).toBe(true);
    expect(canAccessPortalPath(user, '/finance/invoices/new')).toBe(false);
    expect(canAccessPortalPath(user, '/finance/expenses')).toBe(false);
    expect(canAccessPortalPath(user, '/settings/finance')).toBe(false);
  });
  it('keeps Recruit review separate from job management and branding', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('recruit', {
        permissions: ['recruit:candidates:view'],
      }),
    };
    expect(canAccessPortalPath(user, '/recruit/submissions')).toBe(true);
    expect(canAccessPortalPath(user, '/recruit/careers')).toBe(false);
    expect(canAccessPortalPath(user, '/settings/recruit')).toBe(false);
  });
  it('requires a licensed POS grant and terminal management for settings', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('pos', { permissions: ['pos:orders:view'] }),
    };
    expect(canAccessPortalPath(user, '/pos/orders')).toBe(true);
    expect(canAccessPortalPath(user, '/settings/pos')).toBe(false);
    const unlicensed = { effectiveAccess: createEffectiveAccess('pos', { licensed: false }) };
    expect(canAccessPortalPath(unlicensed, '/pos')).toBe(false);
  });
  it('does not turn an HR app grant into general workspace settings access', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('hr', { permissions: ['hr:employees:view'] }),
    };
    expect(canAccessPortalPath(user, '/hr/staff')).toBe(true);
    expect(canAccessPortalPath(user, '/hr/employees/new')).toBe(false);
    expect(canAccessPortalPath(user, '/settings/hr')).toBe(false);
  });
});

describe('standalone Automations access', () => {
  it('allows automation managers at both canonical routes without granting other CRM features', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('crm', { permissions: ['crm:automations:manage'] }),
    };
    expect(canAccessPortalPath(user, '/automations')).toBe(true);
    expect(canAccessPortalPath(user, '/automations/new')).toBe(true);
    expect(canAccessPortalPath(user, '/crm/automations/new')).toBe(true);
    expect(canAccessPortalPath(user, '/crm/customers')).toBe(false);
  });
  it('does not widen access for other app members or unlicensed members', () => {
    expect(
      canAccessPortalPath(
        { effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }) },
        '/automations'
      )
    ).toBe(false);
    expect(
      canAccessPortalPath(
        { effectiveAccess: createEffectiveAccess('crm', { permissions: ['crm:customers:view'] }) },
        '/automations/new'
      )
    ).toBe(false);
    expect(
      canAccessPortalPath(
        {
          effectiveAccess: createEffectiveAccess('crm', {
            permissions: ['crm:automations:manage'],
            licensed: false,
          }),
        },
        '/automations'
      )
    ).toBe(false);
  });
});
