import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ThemeState {
  theme: string;
}

interface ThemeActions {
  toggleTheme: () => void;
  setTheme: (theme: string) => void;
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

      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'theme-store',
    }
  )
);
