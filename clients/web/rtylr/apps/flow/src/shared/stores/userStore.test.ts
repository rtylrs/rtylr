import { describe, it, expect, beforeEach } from 'vitest';
import { useUserStore } from './userStore';
import type { FlowUser } from './userStore';

const { getState, setState } = useUserStore;

const mockUser: FlowUser = {
  id: 'u1',
  name: 'Alice Smith',
  email: 'alice@example.com',
  avatar: 'https://example.com/avatar.png',
  firstName: 'Alice',
};

beforeEach(() => {
  setState({ user: null, token: null, isAuthenticated: false });
});

describe('userStore', () => {
  it('has correct initial state', () => {
    const s = getState();
    expect(s.user).toBeNull();
    expect(s.token).toBeNull();
    expect(s.isAuthenticated).toBe(false);
  });

  it('login sets user, token and isAuthenticated', () => {
    getState().login(mockUser, 'jwt-token-123');
    const s = getState();
    expect(s.user).toEqual(mockUser);
    expect(s.token).toBe('jwt-token-123');
    expect(s.isAuthenticated).toBe(true);
  });

  it('logout clears user, token and isAuthenticated', () => {
    getState().login(mockUser, 'jwt-token-123');
    getState().logout();
    const s = getState();
    expect(s.user).toBeNull();
    expect(s.token).toBeNull();
    expect(s.isAuthenticated).toBe(false);
  });

  it('setUser updates user without affecting token', () => {
    getState().login(mockUser, 'jwt-token-123');
    const updatedUser: FlowUser = { ...mockUser, name: 'Alice Johnson' };
    getState().setUser(updatedUser);
    expect(getState().user?.name).toBe('Alice Johnson');
    expect(getState().token).toBe('jwt-token-123');
  });

  it('setUser works without prior login', () => {
    getState().setUser(mockUser);
    expect(getState().user).toEqual(mockUser);
    expect(getState().isAuthenticated).toBe(false);
  });

  it('login can be called multiple times', () => {
    getState().login(mockUser, 'token-1');
    const otherUser: FlowUser = { id: 'u2', name: 'Bob', email: 'bob@b.com' };
    getState().login(otherUser, 'token-2');
    expect(getState().user?.name).toBe('Bob');
    expect(getState().token).toBe('token-2');
  });

  it('logout after logout is idempotent', () => {
    getState().logout();
    getState().logout();
    expect(getState().isAuthenticated).toBe(false);
  });

  it('user can have optional fields omitted', () => {
    const minimalUser: FlowUser = { id: 'u3', name: 'Min', email: 'min@m.com' };
    getState().login(minimalUser, 'tok');
    expect(getState().user?.avatar).toBeUndefined();
    expect(getState().user?.firstName).toBeUndefined();
  });
});
