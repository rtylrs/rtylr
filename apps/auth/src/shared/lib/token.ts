export interface JwtPayload {
  sub?: string;
  email?: string;
  role?: string;
  exp?: number;
  iat?: number;
  [key: string]: unknown;
}

/**
 * Decodes the payload segment of a JWT without verifying the signature.
 * Returns null for malformed tokens.
 */
export function decodeJwtPayload(token: string): JwtPayload | null {
  if (!token) return null;
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');
    const binary = atob(padded);
    const bytes = Array.from(
      binary,
      (char) => '%' + ('00' + char.charCodeAt(0).toString(16)).slice(-2)
    ).join('');
    const json = decodeURIComponent(bytes);
    return JSON.parse(json) as JwtPayload;
  } catch {
    return null;
  }
}

/**
 * Returns true if the token's exp claim is in the past.
 * Returns false if exp is absent (treat as non-expiring).
 * @param nowMs - current time in milliseconds (defaults to Date.now())
 */
export function isTokenExpired(token: string, nowMs: number = Date.now()): boolean {
  const payload = decodeJwtPayload(token);
  if (!payload || payload.exp === undefined) return false;
  return payload.exp * 1000 < nowMs;
}

/** Extracts the sub claim or returns null. */
export function getTokenSubject(token: string): string | null {
  const payload = decodeJwtPayload(token);
  return typeof payload?.sub === 'string' ? payload.sub : null;
}

/** Extracts the role claim or returns null. */
export function getTokenRole(token: string): string | null {
  const payload = decodeJwtPayload(token);
  return typeof payload?.role === 'string' ? payload.role : null;
}
