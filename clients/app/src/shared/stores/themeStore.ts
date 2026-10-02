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
  document.body.style.background = '';
}

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({
      theme: 'light',
      setTheme: (theme) => {
        applyTheme(theme);
        syncSystemListener(theme);
        set({ theme });
      },
    }),
    { name: 'portal-theme' }
  )
);

let systemListener: { mq: MediaQueryList; handler: () => void } | null = null;

function stopSystemListener() {
  if (systemListener) {
    systemListener.mq.removeEventListener('change', systemListener.handler);
    systemListener = null;
  }
}

/** Follows OS changes only while the resolved theme is `system`. */
function syncSystemListener(theme: Theme) {
  stopSystemListener();
  if (theme !== 'system') return;
  const mq = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = () => applyTheme('system');
  mq.addEventListener('change', handler);
  systemListener = { mq, handler };
}

/** Test hook: whether the OS-change listener is currently attached. */
export function isSystemListenerActive(): boolean {
  return systemListener !== null;
}

/** Call once at app startup before first render to prevent flash */
export function initTheme() {
  try {
    const raw = localStorage.getItem('portal-theme');
    const theme: Theme = raw ? (JSON.parse(raw).state?.theme ?? 'light') : 'light';
    applyTheme(theme);
    syncSystemListener(theme);
  } catch {
    applyTheme('light');
  }
}
