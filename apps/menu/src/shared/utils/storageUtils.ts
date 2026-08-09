/**
 * Browser storage helpers for the Menu app.
 * Wraps localStorage and sessionStorage with safe JSON serialisation.
 */

/** Read and deserialise a value from localStorage. Returns the fallback on error or missing key. */
export function localGet<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

/** Serialise and write a value to localStorage. Returns false if the write fails. */
export function localSet<T>(key: string, value: T): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch {
    return false;
  }
}

/** Remove a key from localStorage. */
export function localRemove(key: string): void {
  localStorage.removeItem(key);
}

/** Read and deserialise a value from sessionStorage. Returns the fallback on error or missing key. */
export function sessionGet<T>(key: string, fallback: T): T {
  try {
    const raw = sessionStorage.getItem(key);
    if (raw === null) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

/** Serialise and write a value to sessionStorage. Returns false if the write fails. */
export function sessionSet<T>(key: string, value: T): boolean {
  try {
    sessionStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch {
    return false;
  }
}

/** Remove a key from sessionStorage. */
export function sessionRemove(key: string): void {
  sessionStorage.removeItem(key);
}
