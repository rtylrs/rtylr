const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Builds a safe `mailto:` href for a member email. Returns null when the
 * value is missing or malformed so callers can render plain text instead of
 * a broken or injected link (query/header injection via `?subject=` or
 * newlines is encoded away, and invalid values never become hrefs).
 */
export function mailtoFor(email: unknown): string | null {
  if (typeof email !== 'string') return null;
  const trimmed = email.trim();
  if (!EMAIL_PATTERN.test(trimmed)) return null;
  return `mailto:${encodeURIComponent(trimmed)}`;
}
