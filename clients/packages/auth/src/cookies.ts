export type CookieSameSite = "lax" | "strict" | "none";

export interface SetCookieOptions {
  days?: number;
  path?: string;
  domain?: string;
  secure?: boolean;
  sameSite?: CookieSameSite;
}

const isBrowser =
  typeof document !== "undefined" && typeof window !== "undefined";

export function getCookie(name: string): string | null {
  if (!isBrowser) return null;
  const nameEQ = `${name}=`;
  const parts = document.cookie.split("; ");
  for (const part of parts) {
    if (part.startsWith(nameEQ)) {
      const raw = part.substring(nameEQ.length);
      try {
        return decodeURIComponent(raw);
      } catch {
        // Malformed percent-encoding must not crash callers (guards run on
        // every route render) - surface the raw value instead.
        return raw;
      }
    }
  }
  return null;
}

export function setCookie(
  name: string,
  value: string,
  options: SetCookieOptions = {},
): void {
  if (!isBrowser) return;

  const {
    days,
    path = "/",
    domain,
    secure = options.secure ?? window.location.protocol === "https:",
    sameSite = "lax",
  } = options;

  let cookie = `${name}=${encodeURIComponent(value)}`;

  if (typeof days === "number") {
    const date = new Date();
    date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
    cookie += `; Expires=${date.toUTCString()}`;
  }

  if (path) cookie += `; Path=${path}`;
  if (domain) cookie += `; Domain=${domain}`;
  if (secure) cookie += `; Secure`;
  if (sameSite) cookie += `; SameSite=${sameSite}`;

  document.cookie = cookie;
}

export function removeCookie(
  name: string,
  options: { path?: string; domain?: string } = {},
): void {
  if (!isBrowser) return;
  const { path = "/", domain } = options;
  let cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Path=${path}`;
  if (domain) cookie += `; Domain=${domain}`;
  document.cookie = cookie;
}

export function getDefaultCookieDomain(): string | undefined {
  if (!isBrowser) return undefined;
  const hostname = window.location.hostname.toLowerCase();

  const isLocalhost =
    hostname === "localhost" || /^\d+\.\d+\.\d+\.\d+$/.test(hostname);
  if (isLocalhost) return undefined;

  if (hostname === "rtylr.com" || hostname.endsWith(".rtylr.com")) {
    return ".rtylr.com";
  }

  return undefined;
}
