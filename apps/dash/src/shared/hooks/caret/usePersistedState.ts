import { useEffect, useReducer } from 'react';
import { readJson, writeJson } from '../../lib/caret/sessionStorage';

/**
 * Like useState but persists into Caret-namespaced localStorage. Identical
 * API surface to useState so call sites can swap freely.
 */
export function usePersistedState<T>(
  key: string,
  initial: T
): [T, React.Dispatch<React.SetStateAction<T>>] {
  type Action = React.SetStateAction<T>;
  const reducer = (state: T, action: Action): T =>
    typeof action === 'function' ? (action as (prev: T) => T)(state) : action;

  const [state, dispatch] = useReducer(reducer, undefined, () => readJson<T>(key) ?? initial);

  useEffect(() => {
    writeJson(key, state);
  }, [key, state]);

  return [state, dispatch];
}
