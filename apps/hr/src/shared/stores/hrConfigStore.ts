import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { HRConfig } from '../types/hrConfig';

interface HRConfigState {
  config: HRConfig | null;
}

interface HRConfigActions {
  setConfig: (config: HRConfig) => void;
  clearConfig: () => void;
}

export const useHRConfigStore = create<HRConfigState & HRConfigActions>()(
  persist(
    (set) => ({
      config: null,
      setConfig: (config: HRConfig) => set({ config }),
      clearConfig: () => set({ config: null }),
    }),
    {
      name: 'rtylr:hr:config',
      partialize: (state) => ({ config: state.config }),
    }
  )
);
