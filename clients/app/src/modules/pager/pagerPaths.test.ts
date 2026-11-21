import { describe, expect, it } from 'vitest';

import { pagerClaimUrl, pagerJoinUrl, portalPathForStaffPage } from './pagerPaths';

describe('portalPathForStaffPage', () => {
  // Each must be a path the Portal's _redirects sends back to the same page.
  it.each([
    ['/board', '/pager'],
    ['/board/', '/pager'],
    ['/display', '/pager/display'],
    ['/poster', '/pager/poster'],
    ['/setup', '/pager/setup'],
  ])('maps %s to %s', (page, portalPath) => {
    expect(portalPathForStaffPage(page)).toBe(portalPath);
  });
});

describe('pagerJoinUrl', () => {
  it('puts the venue at the root of the Rennili origin', () => {
    const url = new URL(pagerJoinUrl('café du port'));
    expect(url.pathname).toBe('/caf%C3%A9%20du%20port');
  });
});

describe('pagerClaimUrl', () => {
  it('comes back to the setup through the Portal, the only origin sign-in returns to', () => {
    const url = new URL(pagerClaimUrl('register'));
    expect(url.pathname).toBe('/register');
    expect(url.searchParams.get('redirect')).toBe('dash');
    expect(url.searchParams.get('returnTo')).toBe('/pager/setup');
  });
});
