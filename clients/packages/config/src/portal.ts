import { PORTAL_URL } from "./apps";

export function resolvePortalReturn(returnTo?: string | null): string {
  if (
    !returnTo ||
    !returnTo.startsWith("/") ||
    returnTo.startsWith("//") ||
    /[\\\u0000-\u0020]/.test(returnTo)
  )
    return PORTAL_URL;
  const target = new URL(returnTo, PORTAL_URL);
  return target.origin === new URL(PORTAL_URL).origin
    ? target.href
    : PORTAL_URL;
}
