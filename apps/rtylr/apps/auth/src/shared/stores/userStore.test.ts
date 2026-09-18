import { describe, it, expect, beforeEach, vi } from 'vitest';

const mockGetCookie = vi.fn(() => null);
const mockRemoveCookie = vi.fn();
const mockSetCookie = vi.fn();
const mockGetDefaultCookieDomain = vi.fn(() => undefined);

vi.mock('@voxire/auth', () => ({
  setCookie: (...args: unknown[]) => mockSetCookie(...args),
  removeCookie: (...args: unknown[]) => mockRemoveCookie(...args),
  getCookie: (...args: unknown[]) => mockGetCookie(...args),
  getDefaultCookieDomain: (...args: unknown[]) => mockGetDefaultCookieDomain(...args),
}));

import { useUserStore } from './userStore';

const { getState } = useUserStore;

beforeEach(() => {
  vi.clearAllMocks();
  useUserStore.setState({
    token: null,
    isAuthenticated: false,
    loading: false,
    error: null,
    role: null,
  });
});

describe('userStore (auth)', () => {
  it('has correct initial state', () => {
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
    expect(getState().role).toBeNull();
  });

  it('login sets role, isAuthenticated, clears loading and error', () => {
    getState().setLoading(true);
    getState().setError('some error');
    getState().login('admin');

    expect(getState().role).toBe('admin');
    expect(getState().isAuthenticated).toBe(true);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
  });

  it('login sets token to null (cookie only)', () => {
    getState().login('owner');
    expect(getState().token).toBeNull();
  });

  it('logout clears all auth state', () => {
    getState().login('admin');

    getState().logout();

    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
    expect(getState().role).toBeNull();
    expect(getState().token).toBeNull();
  });

  it('clearError clears error', () => {
    getState().setError('something went wrong');
    getState().clearError();
    expect(getState().error).toBeNull();
  });

  it('setLoading sets loading state', () => {
    getState().setLoading(true);
    expect(getState().loading).toBe(true);

    getState().setLoading(false);
    expect(getState().loading).toBe(false);
  });

  it('setError sets error message', () => {
    getState().setError('Network error');
    expect(getState().error).toBe('Network error');
  });

  it('setError with null clears error', () => {
    getState().setError('error');
    getState().setError(null);
    expect(getState().error).toBeNull();
  });

  it('getToken calls getCookie', () => {
    mockGetCookie.mockReturnValueOnce('my-token');

    const token = getState().getToken();
    expect(token).toBe('my-token');
    expect(mockGetCookie).toHaveBeenCalledWith('token');
  });

  it('getToken returns null when no cookie', () => {
    const token = getState().getToken();
    expect(token).toBeNull();
  });

  it('logout calls removeCookie', () => {
    getState().login('admin');
    getState().logout();

    expect(mockRemoveCookie).toHaveBeenCalledWith('token', { domain: undefined });
  });
});
