import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { HRCompanyConfigPayload } from '@voxire/voxie/hr/setup/v1/types';

interface HRConfigState {
  config: HRCompanyConfigPayload | null;
}

interface HRConfigActions {
  setConfig: (config: HRCompanyConfigPayload) => void;
  clearConfig: () => void;
}

export const useHRConfigStore = create<HRConfigState & HRConfigActions>()(
  persist(
    (set) => ({
      config: null,
      setConfig: (config: HRCompanyConfigPayload) => set({ config }),
      clearConfig: () => set({ config: null }),
    }),
    {
      name: 'rtylr:hr:config',
      partialize: (state) => ({ config: state.config }),
    }
  )
);
