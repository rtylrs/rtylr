// The API is on a different domain than the app, so the auth cookie can't ride
// cross-site. Like @voxire/api, read the JS-readable `token` cookie and send it
// as a Bearer header — otherwise requests hit the API unauthenticated.
export function readAuthToken(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  const token = match ? decodeURIComponent(match[1]) : null;
  if (!token || token.split('.').length !== 3) return null;
  return token;
}

export function authHeaders(): Record<string, string> {
  const token = readAuthToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}
