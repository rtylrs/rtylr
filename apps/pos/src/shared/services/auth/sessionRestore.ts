type RestoreResult = 'redirected' | 'restoring';

interface RestoreSessionParams {
  cookieToken: string | null;
  storeToken: string | null;
  setToken: (token: string) => void;
  getUserInfo: () => Promise<void>;
  redirectToAuth: () => void;
}

export function restoreUserSessionFromCookie({
  cookieToken,
  storeToken,
  setToken,
  getUserInfo,
  redirectToAuth,
}: RestoreSessionParams): RestoreResult {
  if (!cookieToken) {
    redirectToAuth();
    return 'redirected';
  }

  if (cookieToken !== storeToken) {
    setToken(cookieToken);
  }

  // A rejected restore means the cookie token is expired/revoked: send the
  // terminal back to auth instead of leaving it in limbo until some other
  // API call happens to 401.
  void getUserInfo().catch(() => redirectToAuth());
  return 'restoring';
}
