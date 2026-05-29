import { useState } from 'react';

/**
 * Persist and retrieve a value from sessionStorage, falling back to
 * defaultValue when the key is missing or the stored JSON is invalid.
 */
export function useSessionStorage<T>(
  key: string,
  defaultValue: T
): [T, (value: T) => void, () => void] {
  function read(): T {
    try {
      const raw = sessionStorage.getItem(key);
      if (raw === null) return defaultValue;
      return JSON.parse(raw) as T;
    } catch {
      return defaultValue;
    }
  }

  const [state, setState] = useState<T>(read);

  function write(value: T): void {
    try {
      sessionStorage.setItem(key, JSON.stringify(value));
    } catch {
      // storage quota exceeded or private mode - silently ignore
    }
    setState(value);
  }

  function remove(): void {
    sessionStorage.removeItem(key);
    setState(defaultValue);
  }

  return [state, write, remove];
}
