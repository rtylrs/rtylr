import { useCallback, useMemo, useSyncExternalStore } from 'react';

const KEY = 'flow.lifePreferences.v1';

export interface LifePreferences {
  quietStart: string;
  quietEnd: string;
  /** 0–23, local, preferred hour for a digest summary when reminders exist */
  digestHour: number;
}

const DEFAULTS: LifePreferences = {
  quietStart: '22:00',
  quietEnd: '07:00',
  digestHour: 9,
};

function load(): LifePreferences {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return { ...DEFAULTS };
    const p = JSON.parse(raw) as Partial<LifePreferences>;
    return {
      quietStart: typeof p.quietStart === 'string' ? p.quietStart : DEFAULTS.quietStart,
      quietEnd: typeof p.quietEnd === 'string' ? p.quietEnd : DEFAULTS.quietEnd,
      digestHour:
        typeof p.digestHour === 'number' && p.digestHour >= 0 && p.digestHour <= 23
          ? p.digestHour
          : DEFAULTS.digestHour,
    };
  } catch {
    return { ...DEFAULTS };
  }
}

function save(p: LifePreferences): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(p));
  } catch {
    /* ignore */
  }
}

let cache = load();
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((fn) => fn());
}

export function getLifePreferences(): LifePreferences {
  return cache;
}

export function setLifePreferences(patch: Partial<LifePreferences>): void {
  cache = { ...cache, ...patch };
  save(cache);
  emit();
}

export function useLifePreferences(): {
  prefs: LifePreferences;
  setPrefs: (patch: Partial<LifePreferences>) => void;
} {
  const subscribe = useCallback((onStoreChange: () => void) => {
    listeners.add(onStoreChange);
    return () => listeners.delete(onStoreChange);
  }, []);

  const prefs = useSyncExternalStore(
    subscribe,
    () => cache,
    () => DEFAULTS
  );

  const setPrefs = useCallback((patch: Partial<LifePreferences>) => {
    setLifePreferences(patch);
  }, []);

  return useMemo(() => ({ prefs, setPrefs }), [prefs, setPrefs]);
}
