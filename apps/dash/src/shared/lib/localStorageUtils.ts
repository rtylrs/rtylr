/**
 * Local storage utility helpers for the Dash app.
 * Functions that access storage are side-effecting; pure helpers are noted.
 */

/** Safely get a JSON-parsed value from localStorage. Returns null on any error. */
export function getItem<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

/** Safely set a JSON-serialized value in localStorage. Returns false on error. */
export function setItem<T>(key: string, value: T): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch {
    return false;
  }
}

/** Remove a key from localStorage. Returns false on error. */
export function removeItem(key: string): boolean {
  try {
    localStorage.removeItem(key);
    return true;
  } catch {
    return false;
  }
}

/** Return true when localStorage is available (not blocked by browser policy). */
export function isAvailable(): boolean {
  try {
    const probe = '__dash_ls_probe__';
    localStorage.setItem(probe, '1');
    localStorage.removeItem(probe);
    return true;
  } catch {
    return false;
  }
}

/** Build a namespaced key to avoid collisions across apps. Pure. */
export function namespacedKey(namespace: string, key: string): string {
  return `${namespace}:${key}`;
}

/** Get a value with a fallback default when the key is not found. */
export function getWithDefault<T>(key: string, defaultValue: T): T {
  const stored = getItem<T>(key);
  return stored !== null ? stored : defaultValue;
}

/** Clear all keys matching a given prefix from localStorage. */
export function clearByPrefix(prefix: string): void {
  const keysToRemove: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const k = localStorage.key(i);
    if (k?.startsWith(prefix)) keysToRemove.push(k);
  }
  keysToRemove.forEach((k) => localStorage.removeItem(k));
}
