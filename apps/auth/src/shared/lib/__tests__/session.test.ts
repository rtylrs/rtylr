import { describe, it, expect, vi, afterEach } from 'vitest';
import {
  isSessionValid,
  sessionRemainingMs,
  isSessionExpiringSoon,
  createSession,
  sessionHasRole,
} from '../session';

afterEach(() => {
  vi.useRealTimers();
});

const futureSession = { userId: 'u1', expiresAt: Date.now() + 60_000 };
const expiredSession = { userId: 'u2', expiresAt: Date.now() - 1000 };

describe('isSessionValid', () => {
  it('returns true for a future session', () => {
    expect(isSessionValid(futureSession)).toBe(true);
  });

  it('returns false for an expired session', () => {
    expect(isSessionValid(expiredSession)).toBe(false);
  });
});

describe('sessionRemainingMs', () => {
  it('returns positive ms for valid session', () => {
    expect(sessionRemainingMs(futureSession)).toBeGreaterThan(0);
  });

  it('returns 0 for expired session', () => {
    expect(sessionRemainingMs(expiredSession)).toBe(0);
  });
});

describe('isSessionExpiringSoon', () => {
  it('returns true when session expires within threshold', () => {
    const soon = { userId: 'u', expiresAt: Date.now() + 60_000 };
    expect(isSessionExpiringSoon(soon, 120_000)).toBe(true);
  });

  it('returns false when session has plenty of time', () => {
    const plenty = { userId: 'u', expiresAt: Date.now() + 3_600_000 };
    expect(isSessionExpiringSoon(plenty, 300_000)).toBe(false);
  });

  it('returns false for expired session', () => {
    expect(isSessionExpiringSoon(expiredSession)).toBe(false);
  });
});

describe('createSession', () => {
  it('creates a session with correct userId', () => {
    const s = createSession('user123', 3600_000);
    expect(s.userId).toBe('user123');
  });

  it('creates a session expiring in the future', () => {
    const s = createSession('user123', 3600_000);
    expect(s.expiresAt).toBeGreaterThan(Date.now());
  });

  it('assigns roles when provided', () => {
    const s = createSession('user123', 3600_000, ['admin']);
    expect(s.roles).toEqual(['admin']);
  });
});

describe('sessionHasRole', () => {
  it('returns true when role is present', () => {
    const s = createSession('u', 1000, ['admin', 'editor']);
    expect(sessionHasRole(s, 'admin')).toBe(true);
  });

  it('returns false when role is absent', () => {
    const s = createSession('u', 1000, ['editor']);
    expect(sessionHasRole(s, 'admin')).toBe(false);
  });

  it('returns false when no roles set', () => {
    expect(sessionHasRole({ userId: 'u', expiresAt: Date.now() + 1000 }, 'admin')).toBe(false);
  });
});
