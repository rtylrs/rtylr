import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useUserStore } from './userStore';

// Mock the api module
vi.mock('../services/api/api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
  },
}));

import api from '../services/api/api';

const { getState, setState } = useUserStore;
const act = (fn: () => void) => {
  fn();
};

const initialState = {
  user: null,
  token: null,
  isAuthenticated: false,
  loading: false,
  error: null,
  role: null,
  userId: null,
  addons: {
    feedback: false,
    about: false,
    contact: false,
    landing: false,
  },
  visits: [],
};

beforeEach(() => {
  setState(initialState);
  vi.clearAllMocks();
});

describe('userStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with default unauthenticated state', () => {
    expect(getState().user).toBeNull();
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().loading).toBe(false);
    expect(getState().error).toBeNull();
    expect(getState().role).toBeNull();
    expect(getState().userId).toBeNull();
    expect(getState().addons).toEqual({
      feedback: false,
      about: false,
      contact: false,
      landing: false,
    });
    expect(getState().visits).toEqual([]);
  });

  // ── setToken ───────────────────────────────────────────────────

  it('sets token', () => {
    act(() => getState().setToken('abc123'));
    expect(getState().token).toBe('abc123');
  });

  // ── setUser ────────────────────────────────────────────────────

  it('sets user', () => {
    const user = { name: 'John', email: 'john@test.com' };
    act(() => getState().setUser(user));
    expect(getState().user).toEqual(user);
  });

  // ── logout ─────────────────────────────────────────────────────

  it('logs out and clears state', () => {
    setState({
      user: { name: 'John' },
      token: 'abc',
      isAuthenticated: true,
      role: 'admin',
      userId: '123',
    });

    act(() => getState().logout());

    expect(getState().user).toBeNull();
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().role).toBeNull();
    expect(getState().userId).toBeNull();
  });

  it('logout is safe when already logged out', () => {
    act(() => getState().logout());
    expect(getState().user).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
  });

  // ── clearError ─────────────────────────────────────────────────

  it('clears error', () => {
    setState({ error: 'Something went wrong' });
    act(() => getState().clearError());
    expect(getState().error).toBeNull();
  });

  // ── getUserInfo ────────────────────────────────────────────────

  it('fetches user info successfully', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: {
        userId: 'u1',
        token: 'tok123',
        user: {
          name: 'Alice',
          role: 'admin',
          addons: { feedback: true, about: true, contact: false, landing: false },
        },
        visits: [{ id: 'v1' }],
      },
    });

    await getState().getUserInfo();

    expect(getState().userId).toBe('u1');
    expect(getState().token).toBe('tok123');
    expect(getState().role).toBe('admin');
    expect(getState().isAuthenticated).toBe(true);
    expect(getState().loading).toBe(false);
    expect(getState().addons.feedback).toBe(true);
    expect(getState().visits).toHaveLength(1);
  });

  it('uses nested user id when business profile omits top-level user id', async () => {
    vi.mocked(api.get).mockResolvedValueOnce({
      data: {
        token: 'tok123',
        user: {
          id: 42,
          name: 'Alice',
          role: 'admin',
        },
      },
    });

    await getState().getUserInfo();

    expect(getState().userId).toBe('42');
    expect(getState().isAuthenticated).toBe(true);
  });

  it('sets error on getUserInfo failure', async () => {
    vi.mocked(api.get).mockRejectedValueOnce({
      response: { data: { message: 'Unauthorized' } },
    });

    await expect(getState().getUserInfo()).rejects.toBeTruthy();
    expect(getState().error).toBe('Unauthorized');
    expect(getState().loading).toBe(false);
  });

  it('uses fallback error on getUserInfo failure without message', async () => {
    vi.mocked(api.get).mockRejectedValueOnce(new Error('Network'));

    await expect(getState().getUserInfo()).rejects.toBeTruthy();
    expect(getState().error).toBe('Failed to restore session');
  });

  // ── setThemeAsync ──────────────────────────────────────────────

  it('updates theme on user successfully', async () => {
    setState({ user: { name: 'Alice', theme: 'light' } });
    vi.mocked(api.put).mockResolvedValueOnce({
      data: { theme: { primary: '#000' } },
    });

    await getState().setThemeAsync({ id: 'biz1', themeData: { primary: '#000' } });

    expect(getState().user.theme).toEqual({ primary: '#000' });
    expect(getState().loading).toBe(false);
  });

  it('sets error on setThemeAsync failure', async () => {
    setState({ user: { name: 'Alice' } });
    vi.mocked(api.put).mockRejectedValueOnce({
      response: { data: { message: 'Theme update failed' } },
    });

    await expect(getState().setThemeAsync({ id: 'biz1', themeData: {} })).rejects.toBeTruthy();
    expect(getState().error).toBe('Theme update failed');
    expect(getState().loading).toBe(false);
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('clearError is a no-op when no error', () => {
    act(() => getState().clearError());
    expect(getState().error).toBeNull();
  });

  it('setUser replaces existing user', () => {
    act(() => getState().setUser({ name: 'A' }));
    act(() => getState().setUser({ name: 'B' }));
    expect(getState().user.name).toBe('B');
  });
});
