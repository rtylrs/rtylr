import { describe, it, expect } from 'vitest';
import { PORTAL_URL } from '@voxire/config';

import { authPageLink, signedInDestination } from '../authRedirect';

// The portal sends a signed-out visitor to sign-in with where they were going
// (redirect=dash&returnTo=/flow/...). Switching to sign-up dropped returnTo, so
// a new account landed on the portal home instead of the page they asked for.
describe('authPageLink', () => {
  it('keeps the destination when switching between sign-in and sign-up', () => {
    expect(authPageLink('/register', '?redirect=dash&returnTo=%2Fflow%2Fp%2F5')).toBe(
      '/register?redirect=dash&returnTo=%2Fflow%2Fp%2F5'
    );
    expect(authPageLink('/', '?redirect=dash&returnTo=%2Fflow')).toBe(
      '/?redirect=dash&returnTo=%2Fflow'
    );
  });

  it('carries nothing else across', () => {
    expect(authPageLink('/register', '?redirect=crm&token=abc')).toBe('/register?redirect=crm');
    expect(authPageLink('/register', '')).toBe('/register');
  });
});

describe('signedInDestination', () => {
  it('returns to the portal page the visitor came from', () => {
    expect(signedInDestination('dash', '/flow/p/5')).toBe(new URL('/flow/p/5', PORTAL_URL).href);
  });

  it('never leaves the portal for a return link elsewhere', () => {
    expect(signedInDestination('dash', '//evil.example/x')).toBe(PORTAL_URL);
  });

  it.each(['constructor', 'api', null])('falls back to the portal for %s', (redirect) => {
    expect(signedInDestination(redirect, null)).toBe(PORTAL_URL);
  });
});
