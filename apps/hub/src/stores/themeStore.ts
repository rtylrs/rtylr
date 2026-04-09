import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Theme = 'dark' | 'light' | 'system';

interface ThemeStore {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

function resolveTheme(theme: Theme): 'dark' | 'light' {
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return theme;
}

export function applyTheme(theme: Theme) {
  const resolved = resolveTheme(theme);
  document.documentElement.setAttribute('data-theme', resolved);
  document.body.style.background = resolved === 'dark' ? '#0a0a0a' : '#f4f4f5';
}

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({
      theme: 'dark',
      setTheme: (theme) => {
        applyTheme(theme);
        set({ theme });
      },
    }),
    { name: 'hub-theme' }
  )
);

/** Call once at app startup before first render to prevent flash */
export function initTheme() {
  try {
    const raw = localStorage.getItem('hub-theme');
    const theme: Theme = raw ? (JSON.parse(raw).state?.theme ?? 'dark') : 'dark';
    applyTheme(theme);

    // If system, listen for OS preference changes
    if (theme === 'system') {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        applyTheme('system');
      });
    }
  } catch {
    applyTheme('dark');
  }
}