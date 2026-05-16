import { useEffect, useState } from 'react';
import { readJson, writeJson } from '../../lib/caret/sessionStorage';

/** Mirror a JSON-serialisable value into the Caret-namespaced localStorage. */
export function useLocalStorage<T>(key: string, initial: T): [T, (next: T) => void] {
  const [value, setValue] = useState<T>(() => readJson<T>(key) ?? initial);

  useEffect(() => {
    writeJson(key, value);
  }, [key, value]);

  return [value, setValue];
}
