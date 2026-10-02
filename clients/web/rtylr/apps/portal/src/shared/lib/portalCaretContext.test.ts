import { describe, expect, it } from 'vitest';
import { buildPortalPageContext, buildPortalSystemContext } from './portalCaretContext';

describe('Portal Caret context', () => {
  it('keeps the Insights module and current page when navigation stays inside Portal', () => {
    const input = { pathname: '/insights/settings', businessName: 'Northwind' };
    expect(buildPortalPageContext(input)).toMatchObject({
      app: 'insights',
      page: 'Insights settings',
      businessName: 'Northwind',
      cacheContext: { domains: ['analytics', 'crm', 'erp', 'finance', 'hr', 'pos'] },
    });
    expect(buildPortalSystemContext(input)).toContain('App: RTYLR Insights.');
  });

  it('does not treat team navigation labels as employee records', () => {
    expect(buildPortalPageContext({ pathname: '/team/groups' }).selected).toBeUndefined();
    expect(buildPortalPageContext({ pathname: '/team/invite' }).selected).toBeUndefined();
    expect(buildPortalPageContext({ pathname: '/team/groups/42' }).selected).toEqual({
      group_id: '42',
    });
    expect(buildPortalPageContext({ pathname: '/team/17' }).selected).toEqual({
      employee_id: '17',
    });
  });

  it('retains route filters without putting authentication secrets into the assistant context', () => {
    const context = buildPortalPageContext({
      pathname: '/insights/revenue',
      search: '?period=month&token=private&code=secret',
    });
    expect(context.filters).toEqual({ period: 'month' });
  });
});
