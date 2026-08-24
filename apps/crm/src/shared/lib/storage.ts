/** Reads a value from local storage without allowing storage restrictions to
 * prevent the CRM from booting or rendering. */
export function readLocalStorage(key: string): string | null {
  try {
    if (typeof window === 'undefined') return null;
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

/** Persists a value when local storage is available; otherwise this is a no-op. */
export function writeLocalStorage(key: string, value: string): void {
  try {
    if (typeof window === 'undefined') return;
    window.localStorage.setItem(key, value);
  } catch {
    // Storage can be unavailable in private browsing or when blocked by policy.
  }
}
