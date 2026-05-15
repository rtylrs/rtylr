import { describe, expect, it } from 'vitest';
import { buildDashSystemContext } from './systemContext';

describe('buildDashSystemContext', () => {
  it('uses the page hint for known routes', () => {
    const ctx = buildDashSystemContext({ pathname: '/team' });
    expect(ctx).toContain('Team management');
  });

  it('falls back to a generic hint for unknown paths', () => {
    const ctx = buildDashSystemContext({ pathname: '/some-random' });
    expect(ctx).toContain('navigating the dash app');
  });

  it('matches sub-routes against the prefix', () => {
    const ctx = buildDashSystemContext({ pathname: '/settings/general' });
    expect(ctx).toContain('configuring settings');
  });

  it('includes the business name when provided', () => {
    const ctx = buildDashSystemContext({ pathname: '/', businessName: 'Voxire' });
    expect(ctx).toContain('Business: Voxire');
  });

  it('omits the business prefix when not provided', () => {
    const ctx = buildDashSystemContext({ pathname: '/' });
    expect(ctx.startsWith('User is in the Dash app')).toBe(true);
  });
});
