/** Session state helpers for the auth app. */

export interface Session {
  userId: string;
  expiresAt: number;
  roles?: string[];
}

/**
 * Returns true if the session is still valid (not expired).
 */
export function isSessionValid(session: Session): boolean {
  return Date.now() < session.expiresAt;
}

/**
 * Returns the remaining time in milliseconds for a session, or 0 if expired.
 */
export function sessionRemainingMs(session: Session): number {
  return Math.max(0, session.expiresAt - Date.now());
}

/**
 * Returns true if the session expires within the given threshold in milliseconds.
 */
export function isSessionExpiringSoon(session: Session, thresholdMs = 5 * 60 * 1000): boolean {
  const remaining = sessionRemainingMs(session);
  return remaining > 0 && remaining < thresholdMs;
}

/**
 * Creates a new session object expiring `durationMs` from now.
 */
export function createSession(userId: string, durationMs: number, roles?: string[]): Session {
  return { userId, expiresAt: Date.now() + durationMs, roles };
}

/**
 * Returns true if the session includes the given role.
 */
export function sessionHasRole(session: Session, role: string): boolean {
  return (session.roles ?? []).includes(role);
}
