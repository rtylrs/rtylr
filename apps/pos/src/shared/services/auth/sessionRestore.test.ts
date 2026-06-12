import { describe, expect, it, vi } from 'vitest';
import { restoreUserSessionFromCookie } from './sessionRestore';

describe('restoreUserSessionFromCookie', () => {
  it('redirects to auth when the token cookie is missing', () => {
    const redirectToAuth = vi.fn();

    const result = restoreUserSessionFromCookie({
      cookieToken: null,
      storeToken: null,
      setToken: vi.fn(),
      getUserInfo: vi.fn(),
      redirectToAuth,
    });

    expect(result).toBe('redirected');
    expect(redirectToAuth).toHaveBeenCalledOnce();
  });

  it('hydrates the store token from the cookie before restoring user info', () => {
    const setToken = vi.fn();
    const getUserInfo = vi.fn(() => Promise.resolve());

    const result = restoreUserSessionFromCookie({
      cookieToken: 'cookie-token',
      storeToken: null,
      setToken,
      getUserInfo,
      redirectToAuth: vi.fn(),
    });

    expect(result).toBe('restoring');
    expect(setToken).toHaveBeenCalledWith('cookie-token');
    expect(getUserInfo).toHaveBeenCalledOnce();
  });

  it('restores user info without rewriting an unchanged token', () => {
    const setToken = vi.fn();
    const getUserInfo = vi.fn(() => Promise.resolve());

    const result = restoreUserSessionFromCookie({
      cookieToken: 'store-token',
      storeToken: 'store-token',
      setToken,
      getUserInfo,
      redirectToAuth: vi.fn(),
    });

    expect(result).toBe('restoring');
    expect(setToken).not.toHaveBeenCalled();
    expect(getUserInfo).toHaveBeenCalledOnce();
  });
});
