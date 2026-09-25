import 'fake-indexeddb/auto';
import { afterEach, beforeEach } from 'vitest';

function createMemoryStorage(): Storage {
  const entries = new Map<string, string>();

  return {
    get length() {
      return entries.size;
    },
    clear() {
      entries.clear();
    },
    getItem(key: string) {
      return entries.has(key) ? entries.get(key)! : null;
    },
    key(index: number) {
      return Array.from(entries.keys())[index] ?? null;
    },
    removeItem(key: string) {
      entries.delete(key);
    },
    setItem(key: string, value: string) {
      entries.set(key, String(value));
    },
  };
}

function ensureStorage(name: 'localStorage' | 'sessionStorage') {
  if (typeof window === 'undefined') return;

  try {
    if (window[name]) return;
  } catch {
    // Fall through and define a test storage below.
  }

  Object.defineProperty(window, name, {
    configurable: true,
    value: createMemoryStorage(),
  });
}

function clearCookies() {
  if (typeof document === 'undefined') return;

  const names = new Set(
    document.cookie
      .split(';')
      .map((cookie) => cookie.split('=')[0]?.trim())
      .filter((name): name is string => Boolean(name))
  );
  for (const name of names) {
    document.cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0`;
    document.cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; path=/`;
    document.cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; Path=/`;
  }
}

ensureStorage('localStorage');
ensureStorage('sessionStorage');

beforeEach(() => {
  ensureStorage('localStorage');
  ensureStorage('sessionStorage');
  window.localStorage.clear();
  window.sessionStorage.clear();
  clearCookies();
});

afterEach(() => {
  window.localStorage.clear();
  window.sessionStorage.clear();
  clearCookies();
});
