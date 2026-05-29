import { describe, it, expect } from 'vitest';
import { decodeJwtPayload, isTokenExpired, getTokenSubject, getTokenRole } from '../token';

function buildFakeJwt(payload: Record<string, unknown>): string {
  const enc = (obj: unknown) =>
    btoa(JSON.stringify(obj)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `${enc({ alg: 'HS256', typ: 'JWT' })}.${enc(payload)}.fakesig`;
}

const futureExp = Math.floor(Date.now() / 1000) + 3600;
const pastExp = Math.floor(Date.now() / 1000) - 3600;

describe('decodeJwtPayload', () => {
  it('decodes a valid JWT payload', () => {
    const token = buildFakeJwt({ sub: 'user-1', role: 'owner' });
    const payload = decodeJwtPayload(token);
    expect(payload?.sub).toBe('user-1');
    expect(payload?.role).toBe('owner');
  });

  it('returns null for a malformed token (not 3 parts)', () => {
    expect(decodeJwtPayload('bad.token')).toBeNull();
  });

  it('returns null for empty string', () => {
    expect(decodeJwtPayload('')).toBeNull();
  });
});

describe('isTokenExpired', () => {
  it('returns false for a token with a future exp', () => {
    expect(isTokenExpired(buildFakeJwt({ exp: futureExp }))).toBe(false);
  });

  it('returns true for a token with a past exp', () => {
    expect(isTokenExpired(buildFakeJwt({ exp: pastExp }))).toBe(true);
  });

  it('returns false when exp is absent', () => {
    expect(isTokenExpired(buildFakeJwt({ sub: 'user-1' }))).toBe(false);
  });

  it('respects a custom nowMs argument', () => {
    const token = buildFakeJwt({ exp: 1000 });
    expect(isTokenExpired(token, 500_000)).toBe(false);
    expect(isTokenExpired(token, 2_000_000)).toBe(true);
  });
});

describe('getTokenSubject', () => {
  it('extracts the sub claim', () => {
    expect(getTokenSubject(buildFakeJwt({ sub: 'user-42' }))).toBe('user-42');
  });

  it('returns null when sub is absent', () => {
    expect(getTokenSubject(buildFakeJwt({ role: 'admin' }))).toBeNull();
  });
});

describe('getTokenRole', () => {
  it('extracts the role claim', () => {
    expect(getTokenRole(buildFakeJwt({ role: 'admin' }))).toBe('admin');
  });

  it('returns null when role is absent', () => {
    expect(getTokenRole(buildFakeJwt({ sub: 'user-1' }))).toBeNull();
  });
});
