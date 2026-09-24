type RestoreResult = "redirected" | "restoring";

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
    return "redirected";
  }

  if (cookieToken !== storeToken) {
    setToken(cookieToken);
  }

  void getUserInfo().catch((error: unknown) => {
    const status = httpStatus(error);
    if (status === 401 || status === 403) {
      redirectToAuth();
    }
  });
  return "restoring";
}

function httpStatus(error: unknown): number | undefined {
  if (!error || typeof error !== "object") return undefined;
  const value = error as {
    status?: unknown;
    statusCode?: unknown;
    response?: { status?: unknown; statusCode?: unknown };
  };
  const status =
    value.statusCode ??
    value.status ??
    value.response?.status ??
    value.response?.statusCode;
  return typeof status === "number" ? status : undefined;
}
