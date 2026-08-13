import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Theme = 'dark' | 'light' | 'system';

/**
 * Dark, because that is the mode Flow is designed in: every screen in the design
 * set is drawn dark, and a bare `:root` in globals.css resolves to the dark
 * palette. Light remains a first-class opt-in, and a stored preference always
 * wins, so anyone who already chose light keeps it.
 */
export const DEFAULT_THEME: Theme = 'dark';

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
  } else {
    root.setAttribute('data-theme', theme);
  }
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: DEFAULT_THEME,
      setTheme: (theme) => {
        applyTheme(theme);
        set({ theme });
      },
    }),
    { name: 'flow-theme' }
  )
);

export function initTheme() {
  const stored = localStorage.getItem('flow-theme');
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      if (parsed.state?.theme) {
        applyTheme(parsed.state.theme);
        return;
      }
    } catch {}
  }
  applyTheme(DEFAULT_THEME);
}
