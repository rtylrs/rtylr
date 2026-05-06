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

  void getUserInfo().catch(() => {});
  return 'restoring';
}
