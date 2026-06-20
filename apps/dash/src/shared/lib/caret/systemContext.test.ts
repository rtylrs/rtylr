import { describe, expect, it } from 'vitest';
import { buildDashPageContext, buildDashSystemContext } from './systemContext';

describe('buildDashSystemContext', () => {
  it('uses the page hint for known routes', () => {
    const ctx = buildDashSystemContext({ pathname: '/team' });
    expect(ctx).toContain('Team management');
    expect(ctx).toContain('Current module: team');
  });

  it('falls back to a generic hint for unknown paths', () => {
    const ctx = buildDashSystemContext({ pathname: '/some-random' });
    expect(ctx).toContain('navigating the dash app');
  });

  it('matches sub-routes against the prefix', () => {
    const ctx = buildDashSystemContext({ pathname: '/settings/general' });
    expect(ctx).toContain('business settings');
  });

  it('includes the business name when provided', () => {
    const ctx = buildDashSystemContext({ pathname: '/', businessName: 'Voxire' });
    expect(ctx).toContain('Business: Voxire');
  });

  it('omits the business prefix when not provided', () => {
    const ctx = buildDashSystemContext({ pathname: '/' });
    expect(ctx.startsWith('App: Dash business command center')).toBe(true);
  });

  it('builds structured page context', () => {
    const ctx = buildDashPageContext({
      pathname: '/team/42',
      businessName: 'Voxire',
      role: 'owner',
      apps: ['pos', 'finance'],
    });

    expect(ctx.app).toBe('dash');
    expect(ctx.module).toBe('team');
    expect(ctx.businessName).toBe('Voxire');
    expect(ctx.role).toBe('owner');
    expect(ctx.selected).toEqual({ employee_id: '42' });
    expect(ctx.enabledApps).toEqual(['pos', 'finance']);
    expect(ctx.businessMemory?.currency).toBe('USD');
    expect(ctx.cacheContext?.domains).toEqual(['hr', 'team']);
    expect(ctx.dataFreshness).toEqual(ctx.cacheContext);
  });
});
