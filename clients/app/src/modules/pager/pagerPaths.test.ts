import { describe, expect, it } from 'vitest';

import {
  pagerClaimUrl,
  pagerJoinUrl,
  pagerSignedOutUrl,
  portalPathForStaffPage,
} from './pagerPaths';

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

describe('a setup link for a new venue', () => {
  const search = '?name=Brio+Coffee&slug=brio&email=owner%40brio.example';

  it('signs a visitor up with the name and email filled in, and keeps the slug for the setup', () => {
    const url = new URL(pagerSignedOutUrl('/setup', search) ?? '');
    expect(url.pathname).toBe('/register');
    expect(url.searchParams.get('businessName')).toBe('Brio Coffee');
    expect(url.searchParams.get('email')).toBe('owner@brio.example');
    const returnTo = url.searchParams.get('returnTo') ?? '';
    expect(returnTo).toBe('/pager/setup?name=Brio+Coffee&slug=brio&email=owner%40brio.example');
    // The sign-in app refuses a returnTo with spaces or control characters.
    expect(returnTo).not.toMatch(/[\s\u0000-\u001f]/);
  });

  it('leaves every other signed-out visit to the sign-in page', () => {
    expect(pagerSignedOutUrl('/setup', '')).toBeNull();
    expect(pagerSignedOutUrl('/setup', '?utm_source=x')).toBeNull();
    expect(pagerSignedOutUrl('/board', search)).toBeNull();
  });
});
