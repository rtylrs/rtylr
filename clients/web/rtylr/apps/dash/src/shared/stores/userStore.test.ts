import { describe, it, expect, beforeEach } from 'vitest';
import { useUserStore } from './userStore';

const { getState } = useUserStore;

beforeEach(() => {
  // Reset to initial state
  useUserStore.setState({
    user: null,
    token: null,
    isAuthenticated: false,
    userId: null,
    addons: { feedback: false, about: false, contact: false, landing: false },
    visits: [],
  });
});

describe('userStore (dash)', () => {
  it('has correct initial state', () => {
    expect(getState().user).toBeNull();
    expect(getState().token).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().userId).toBeNull();
    expect(getState().visits).toEqual([]);
  });

  it('has correct initial addons', () => {
    const { addons } = getState();
    expect(addons.feedback).toBe(false);
    expect(addons.about).toBe(false);
    expect(addons.contact).toBe(false);
    expect(addons.landing).toBe(false);
  });

  it('login sets user and isAuthenticated', () => {
    const user = { fullName: 'John', email: 'john@test.com' };
    getState().login(user);

    expect(getState().user).toEqual(user);
    expect(getState().isAuthenticated).toBe(true);
  });

  it('logout clears user state', () => {
    getState().login({ fullName: 'John' });
    getState().setToken('abc');
    getState().setUserId('u1');

    getState().logout();

    expect(getState().user).toBeNull();
    expect(getState().isAuthenticated).toBe(false);
    expect(getState().token).toBeNull();
    expect(getState().userId).toBeNull();
  });

  it('setToken sets token and authenticates', () => {
    getState().setToken('my-token');

    expect(getState().token).toBe('my-token');
    expect(getState().isAuthenticated).toBe(true);
  });

  it('setToken with empty string sets isAuthenticated false', () => {
    getState().setToken('');

    expect(getState().token).toBe('');
    expect(getState().isAuthenticated).toBe(false);
  });

  it('setUser updates user', () => {
    const user = { id: 1, fullName: 'Test' };
    getState().setUser(user);
    expect(getState().user).toEqual(user);
  });

  it('setUserId updates userId', () => {
    getState().setUserId('user-123');
    expect(getState().userId).toBe('user-123');
  });
});
