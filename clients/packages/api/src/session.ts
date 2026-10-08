/**
 * Shared-session invalidation for the SSO cookie.
 *
 * The `token` cookie is written against the parent domain, so all ten SPAs read
 * and clear the same one. Any app that clears it signs the user out of every
 * other app too, which is why 401 handling has to decide carefully rather than
 * react to the status code.
 *
 * Four things are deliberately NOT reasons to end the shared session:
 *
 * 1. A request that carried no credential. `auth_middleware.go` answers every
 *    authenticated route with 401 "Unauthorized" when the `Authorization`
 *    header is absent, so an anonymous probe, a public endpoint, or a request
 *    issued before the token cookie was readable all land here with a perfectly
 *    good session in the other nine apps.
 * 2. A 401 whose body names a non-session cause. The backend reuses 401 for
 *    "Email verification required", "Account activation required", "MFA
 *    enrollment required", "MFA verification required" and "Authorization
 *    unavailable" (a failed lookup, not an invalid token) - all emitted for a
 *    perfectly valid token.
 * 3. A rejected token that is no longer the current one. A stale background
 *    poll carries the previous token; a sign-in elsewhere rewrote the cookie
 *    before the 401 came back. Clearing now would destroy the live session.
 * 4. Any 401 while a refresh is in flight. Concurrent 401s during a refresh are
 *    the race itself, not evidence of a dead session.
 */

/**
 * Credential each request carried, keyed on the config object the interceptor
 * already passes to both hooks (axios hands the same object back as
 * `error.config`). A WeakMap keeps this off the config itself: the request
 * config is Axios's to shape, and a symbol-keyed property written from here is
 * invisible to its types, so reading it back needs an unchecked assertion.
 */
const credentials = new WeakMap<object, string | null>();

/** Record the credential a request is carrying, for the 401 handler to read. */
export function markCredential(config: unknown, token: string | null): void {
  if (!config || typeof config !== 'object') return;
  credentials.set(config, token);
}

/** The credential a request carried, or null when it sent none. */
export function sentCredential(config: unknown): string | null {
  if (!config || typeof config !== 'object') return null;
  return credentials.get(config) ?? null;
}

/** The result of checking whether a refresh token can renew the session. */
export type SessionRefreshResult = 'refreshed' | 'rejected' | 'unavailable';

/** The session check a 401 is allowed to run before any sign-out. */
export type SessionRefresh = (signal: AbortSignal) => Promise<SessionRefreshResult>;

const sessionRefreshTimeoutMs = 5_000;

let sessionRefresh: SessionRefresh | null = null;
let refreshInFlight: Promise<SessionRefreshResult> | null = null;

/**
 * Token whose 401 already ended the shared session. Keyed on the token rather
 * than a boolean so signing back in re-arms it: a session established after the
 * sign-out must still be able to end on its own first 401.
 */
let invalidatedToken: string | null = null;

export function setSessionRefresh(refresh: SessionRefresh | null): void {
  sessionRefresh = refresh;
  refreshInFlight = null;
}

/** Test seam: forget the recorded sign-out without touching the cookie. */
export function resetSessionInvalidationForTests(): void {
  invalidatedToken = null;
  sessionRefresh = null;
  refreshInFlight = null;
}

/**
 * Run the registered session check, collapsing concurrent callers onto one
 * attempt. Always settles - a check that throws counts as "still signed in",
 * because a failed check is not evidence that the session died, and the 401
 * itself stays the caller's to handle.
 */
function refreshSession(): Promise<SessionRefreshResult> {
  if (!sessionRefresh) return Promise.resolve('rejected');
  if (refreshInFlight) return refreshInFlight;

  const controller = new AbortController();
  let timeout: ReturnType<typeof setTimeout> | undefined;
  const timedOut = new Promise<SessionRefreshResult>((resolve) => {
    timeout = setTimeout(() => {
      controller.abort();
      resolve('unavailable');
    }, sessionRefreshTimeoutMs);
  });

  const attempt = (async (): Promise<SessionRefreshResult> => {
    try {
      return await Promise.race([sessionRefresh!(controller.signal), timedOut]);
    } catch {
      return 'unavailable';
    } finally {
      // Cleared in `finally` so a throw cannot wedge every later 401 behind a
      // settled promise nobody awaits.
      if (timeout !== undefined) clearTimeout(timeout);
      controller.abort();
      refreshInFlight = null;
    }
  })();

  refreshInFlight = attempt;
  return attempt;
}

/** Read the live token, treating an unreadable cookie jar as "unknown". */
function safeRead(readToken: () => string | null): string | null | undefined {
  try {
    return readToken();
  } catch {
    // A cookie jar that cannot be read is not evidence the session is dead.
    return undefined;
  }
}

/**
 * Decide whether a 401 may end the shared session, then report it.
 *
 * `token` is the credential the request actually carried (null when it carried
 * none) and `readToken` re-reads the live cookie. `endSession` performs the
 * cookie clear and the redirect, so this module never touches the DOM itself.
 *
 * Never rejects: the axios path fires this without awaiting, so a throw here
 * would surface as an unhandled rejection and hide the 401 the caller is
 * handling. An unreadable cookie stops the sign-out rather than forcing one.
 */
export async function shouldEndSharedSession(
  token: string | null,
  isSessionRejection: boolean,
  readToken: () => string | null,
  endSession: () => void
): Promise<boolean> {
  try {
    if (!token) return false;
    if (!isSessionRejection) return false;
    if (safeRead(readToken) !== token) return false;

    if (sessionRefresh) {
      const refreshResult = await refreshSession();
      // A successful refresh replaced the cookie, so this 401 belonged to the
      // previous token and there is nothing left to end.
      if (refreshResult !== 'rejected') return false;
      if (safeRead(readToken) !== token) return false;
    }

    if (invalidatedToken === token) return false;
    invalidatedToken = token;
    endSession();
    return true;
  } catch {
    return false;
  }
}
