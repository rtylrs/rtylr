import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface HRSetupState {
  currentStepIndex: number;
}

interface HRSetupActions {
  setCurrentStepIndex: (index: number) => void;
  reset: () => void;
}

export const useHRSetupStore = create<HRSetupState & HRSetupActions>()(
  persist(
    (set) => ({
      currentStepIndex: 0,
      setCurrentStepIndex: (currentStepIndex: number) => set({ currentStepIndex }),
      reset: () => set({ currentStepIndex: 0 }),
    }),
    {
      name: 'rtylr:hr:setup',
      partialize: (state) => ({ currentStepIndex: state.currentStepIndex }),
    }
  )
);
