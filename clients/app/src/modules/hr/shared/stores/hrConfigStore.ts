import { create } from 'zustand';
import type { HRCompanyConfigPayload } from '@voxire/voxie/hr/setup/v1/types';

interface HRConfigState {
  config: HRCompanyConfigPayload | null;
}

interface HRConfigActions {
  setConfig: (config: HRCompanyConfigPayload) => void;
  clearConfig: () => void;
}

export const useHRConfigStore = create<HRConfigState & HRConfigActions>()((set) => ({
  config: null,
  setConfig: (config) => set({ config }),
  clearConfig: () => set({ config: null }),
}));
