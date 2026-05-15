/**
 * Tiny wrapper around localStorage scoped to the Caret namespace. Keeps
 * key names consistent and survives quota or disabled-storage errors.
 */

const NS = 'caret:';

export function readJson<T>(key: string): T | null {
  if (typeof localStorage === 'undefined') return null;
  try {
    const raw = localStorage.getItem(NS + key);
    if (!raw) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

export function writeJson<T>(key: string, value: T): void {
  if (typeof localStorage === 'undefined') return;
  try {
    localStorage.setItem(NS + key, JSON.stringify(value));
  } catch {
    // ignore — caller treats persistence as best effort
  }
}

export function removeKey(key: string): void {
  if (typeof localStorage === 'undefined') return;
  try {
    localStorage.removeItem(NS + key);
  } catch {
    // ignore
  }
}
