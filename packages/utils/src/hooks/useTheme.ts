import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type ThemeMode = 'light' | 'dark' | 'system';

interface ThemeStore {
  mode: ThemeMode;
  resolved: 'light' | 'dark';
  setMode: (mode: ThemeMode) => void;
  toggle: () => void;
}

function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
  if (mode === 'system') {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return mode;
}

/**
 * Creates a Zustand theme store for a specific app.
 * Persists to localStorage under the given key.
 *
 * Usage:
 * ```ts
 * export const useThemeStore = createThemeStore('rtylr-store-theme');
 * ```
 */
export function createThemeStore(storageKey: string, defaultMode: ThemeMode = 'light') {
  return create<ThemeStore>()(
    persist(
      (set) => ({
        mode: defaultMode,
        resolved: resolveTheme(defaultMode),
        setMode: (mode) => set({ mode, resolved: resolveTheme(mode) }),
        toggle: () =>
          set((s) => {
            const next = s.resolved === 'dark' ? 'light' : 'dark';
            return { mode: next, resolved: next };
          }),
      }),
      {
        name: storageKey,
        onRehydrateStorage: () => (state) => {
          if (state) {
            state.resolved = resolveTheme(state.mode);
          }
        },
      }
    )
  );
}

/**
 * Apply the resolved theme to the document root.
 * Call this in your app's layout or main entry.
 */
export function applyTheme(resolved: 'light' | 'dark') {
  document.documentElement.setAttribute('data-theme', resolved);
  document.documentElement.classList.toggle('dark', resolved === 'dark');
}
