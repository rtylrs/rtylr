/** Gets a value from localStorage, returning null if not available. */
export function getItem<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

/** Sets a value in localStorage. Returns true on success, false on failure. */
export function setItem<T>(key: string, value: T): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch {
    return false;
  }
}

/** Removes a key from localStorage. */
export function removeItem(key: string): void {
  try {
    localStorage.removeItem(key);
  } catch {
    // ignore
  }
}

/** Returns true if a key exists in localStorage. */
export function hasItem(key: string): boolean {
  try {
    return localStorage.getItem(key) !== null;
  } catch {
    return false;
  }
}

/** Gets a value from sessionStorage, returning null if not available. */
export function getSessionItem<T>(key: string): T | null {
  try {
    const raw = sessionStorage.getItem(key);
    if (raw === null) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

/** Sets a value in sessionStorage. Returns true on success, false on failure. */
export function setSessionItem<T>(key: string, value: T): boolean {
  try {
    sessionStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch {
    return false;
  }
}

/** Flow app localStorage key prefix. */
export const FLOW_STORAGE_PREFIX = 'flow:';

/** Returns a namespaced storage key for the flow app. */
export function flowKey(key: string): string {
  return `${FLOW_STORAGE_PREFIX}${key}`;
}
