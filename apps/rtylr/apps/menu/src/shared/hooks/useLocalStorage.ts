import { useCallback, useEffect, useState } from 'react';

/**
 * Reads the initial value from localStorage, falling back to the provided
 * default. Keeps the stored value and state in sync. Returns a setter that
 * persists the new value to localStorage.
 */
export function useLocalStorage<T>(key: string, defaultValue: T): [T, (value: T) => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = localStorage.getItem(key);
      return item !== null ? (JSON.parse(item) as T) : defaultValue;
    } catch {
      return defaultValue;
    }
  });

  const setValue = useCallback(
    (value: T) => {
      try {
        setStoredValue(value);
        localStorage.setItem(key, JSON.stringify(value));
      } catch {
        // Silently ignore storage quota errors.
      }
    },
    [key]
  );

  useEffect(() => {
    function onStorage(e: StorageEvent) {
      if (e.key !== key) return;
      try {
        const next = e.newValue !== null ? (JSON.parse(e.newValue) as T) : defaultValue;
        setStoredValue(next);
      } catch {
        // Ignore parse errors from cross-tab changes.
      }
    }
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  }, [key, defaultValue]);

  return [storedValue, setValue];
}
