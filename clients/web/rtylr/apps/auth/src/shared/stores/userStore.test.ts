import { describe, it, expect, beforeEach, vi } from 'vitest';

const { mockGetCookie, mockRemoveCookie, mockSetCookie, mockGetDefaultCookieDomain } = vi.hoisted(
  () => ({
    mockGetCookie: vi.fn<(name: string) => string | null>(() => null),
    mockRemoveCookie: vi.fn(),
    mockSetCookie: vi.fn(),
    mockGetDefaultCookieDomain: vi.fn<() => string | undefined>(() => undefined),
  })
);

vi.mock('@voxire/auth', () => ({
  setCookie: mockSetCookie,
  removeCookie: mockRemoveCookie,
  getCookie: mockGetCookie,
  getDefaultCookieDomain: mockGetDefaultCookieDomain,
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
  });
});

describe('userStore (auth)', () => {
  it('has correct initial state', () => {
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
  });

  it('login sets isAuthenticated and clears loading and error', () => {
    getState().setLoading(true);
    getState().setError('some error');
    getState().login();

    expect(getState().isAuthenticated).toBe(true);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
  });

  it('login sets token to null (cookie only)', () => {
    getState().login();
    expect(getState().token).toBeNull();
  });

  it('logout clears all auth state', () => {
    getState().login();

    getState().logout();

    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
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
    getState().login();
    getState().logout();

    expect(mockRemoveCookie).toHaveBeenCalledWith('token', { domain: undefined });
  });
});
