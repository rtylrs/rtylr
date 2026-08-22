import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type BusinessType =
  'restaurant' | 'gym' | 'retail' | 'salon' | 'cafe' | 'bar' | 'hotel' | 'other';

interface BusinessState {
  businessType: BusinessType | null;
  businessPreferences: Record<string, unknown>;
}

interface BusinessActions {
  setBusinessType: (type: BusinessType) => void;
  setBusinessPreferences: (preferences: Record<string, unknown>) => void;
  updateBusinessPreference: (key: string, value: unknown) => void;
  reset: () => void;
}

const initialState: BusinessState = {
  businessType: null,
  businessPreferences: {},
};

export const useBusinessStore = create<BusinessState & BusinessActions>()(
  persist(
    (set) => ({
      ...initialState,

      setBusinessType: (type) => set({ businessType: type }),

      setBusinessPreferences: (preferences) => set({ businessPreferences: preferences }),

      updateBusinessPreference: (key, value) =>
        set((state) => ({
          businessPreferences: {
            ...state.businessPreferences,
            [key]: value,
          },
        })),

      reset: () => set(initialState),
    }),
    {
      name: 'business-store',
      partialize: (state) => ({
        businessType: state.businessType,
        businessPreferences: state.businessPreferences,
      }),
    }
  )
);
