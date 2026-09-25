import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type POSTheme = 'dark' | 'light' | 'midnight' | 'slate' | 'aurora';

export const POS_THEME_OPTIONS: Array<{ value: POSTheme; label: string }> = [
  { value: 'dark', label: 'Landing dark' },
  { value: 'light', label: 'Light' },
  { value: 'midnight', label: 'Midnight' },
  { value: 'slate', label: 'Slate' },
  { value: 'aurora', label: 'Aurora' },
];

interface ThemeState {
  theme: POSTheme;
}

interface ThemeActions {
  toggleTheme: () => void;
  setTheme: (theme: string) => void;
}

export function normalizePOSTheme(theme: unknown): POSTheme {
  return POS_THEME_OPTIONS.some((option) => option.value === theme) ? (theme as POSTheme) : 'dark';
}

function mergeTheme(
  persistedState: unknown,
  currentState: ThemeState & ThemeActions
): ThemeState & ThemeActions {
  const persisted = (persistedState ?? {}) as Partial<ThemeState>;
  return {
    ...currentState,
    ...persisted,
    theme: normalizePOSTheme(persisted.theme ?? currentState.theme),
  };
}

const initialState: ThemeState = {
  theme: 'dark',
};

export const useThemeStore = create<ThemeState & ThemeActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      toggleTheme: () => {
        const currentTheme = get().theme;
        set({ theme: currentTheme === 'light' ? 'dark' : 'light' });
      },

      setTheme: (theme) => set({ theme: normalizePOSTheme(theme) }),
    }),
    {
      name: 'theme-store',
      merge: mergeTheme,
    }
  )
);
